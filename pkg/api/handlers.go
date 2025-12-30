package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"ksight/pkg/service"
	"ksight/pkg/service/scheduler"
)

type clusterService interface {
	AddCluster(name, kubeconfig, context string) (string, error)
	RemoveCluster(clusterID string) error
	GetClusters() map[string]service.ClusterInfo
	AddResourceWatcher(request service.ResourceWatchRequest) error
	RemoveResourceWatcher(request service.ResourceWatchRequest) error
	LoadInitialData(clusterID string, group, version, resource string) ([]map[string]any, string, error)
	LoadKubeconfigFromFile(filePath string) (string, error)
}

type schedulerAggregator interface {
	StartAggregation(request scheduler.SchedulerAggregationRequest) error
	StopAggregation(clusterID string) error
	GetSnapshot(clusterID string) (*scheduler.SchedulerSnapshot, error)
}

// Handler holds dependencies for HTTP handlers
type Handler struct {
	clusterService      clusterService
	schedulerAggregator schedulerAggregator
	wsHub               *WebSocketHub
}

// NewHandler creates a new API handler
func NewHandler(cs clusterService, sa schedulerAggregator, hub *WebSocketHub) *Handler {
	return &Handler{
		clusterService:      cs,
		schedulerAggregator: sa,
		wsHub:               hub,
	}
}

// GetClusters returns all clusters
func (h *Handler) GetClusters(w http.ResponseWriter, r *http.Request) {
	clusters := h.clusterService.GetClusters()
	writeJSON(w, http.StatusOK, clusters)
}

// AddClusterRequest represents the request body for adding a cluster
type AddClusterRequest struct {
	Name       string `json:"name"`
	Kubeconfig string `json:"kubeconfig"`
	Context    string `json:"context"`
}

// AddCluster adds a new cluster
func (h *Handler) AddCluster(w http.ResponseWriter, r *http.Request) {
	var req AddClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	clusterID, err := h.clusterService.AddCluster(req.Name, req.Kubeconfig, req.Context)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"clusterId": clusterID})
}

// RemoveCluster removes a cluster
func (h *Handler) RemoveCluster(w http.ResponseWriter, r *http.Request) {
	clusterID := r.PathValue("id")
	if clusterID == "" {
		writeError(w, http.StatusBadRequest, "Cluster ID required")
		return
	}

	if err := h.clusterService.RemoveCluster(clusterID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// StartSchedulerAggregation starts scheduler aggregation for a cluster
func (h *Handler) StartSchedulerAggregation(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	var req scheduler.SchedulerAggregationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	log.Printf("Starting scheduler aggregation for cluster: %s", req.ClusterID)

	if err := h.schedulerAggregator.StartAggregation(req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// StopSchedulerAggregation stops scheduler aggregation for a cluster
func (h *Handler) StopSchedulerAggregation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClusterID string `json:"clusterId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.schedulerAggregator.StopAggregation(req.ClusterID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// GetSchedulerSnapshot returns the current scheduler snapshot
func (h *Handler) GetSchedulerSnapshot(w http.ResponseWriter, r *http.Request) {
	clusterID := r.PathValue("clusterId")
	if clusterID == "" {
		writeError(w, http.StatusBadRequest, "Cluster ID required")
		return
	}

	snapshot, err := h.schedulerAggregator.GetSnapshot(clusterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, snapshot)
}

// HandleWebSocket upgrades the connection to WebSocket
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ServeWebSocket(h.wsHub, w, r)
}

// GetNodes returns all nodes for a cluster
func (h *Handler) GetNodes(w http.ResponseWriter, r *http.Request) {
	clusterID := r.PathValue("clusterId")
	if clusterID == "" {
		writeError(w, http.StatusBadRequest, "Cluster ID required")
		return
	}

	// Ensure we have a watcher for nodes
	_ = h.clusterService.AddResourceWatcher(service.ResourceWatchRequest{
		ClusterID: clusterID,
		Group:     "",
		Version:   "v1",
		Resource:  "nodes",
	})

	// Load node data
	nodes, _, err := h.clusterService.LoadInitialData(clusterID, "", "v1", "nodes")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Transform nodes to a simpler format for the frontend
	var result []map[string]any
	for _, node := range nodes {
		metadata, _ := node["metadata"].(map[string]any)
		spec, _ := node["spec"].(map[string]any)
		status, _ := node["status"].(map[string]any)

		if metadata == nil {
			continue
		}

		name, _ := metadata["name"].(string)
		labels, _ := metadata["labels"].(map[string]any)
		creationTimestamp, _ := metadata["creationTimestamp"].(string)

		// Extract roles from labels
		var roles []string
		for key := range labels {
			if len(key) > 24 && key[:24] == "node-role.kubernetes.io/" {
				roles = append(roles, key[24:])
			}
		}
		rolesStr := "-"
		if len(roles) > 0 {
			rolesStr = ""
			for i, role := range roles {
				if i > 0 {
					rolesStr += ", "
				}
				rolesStr += role
			}
		}

		// Extract version from nodeInfo
		version := "-"
		if nodeInfo, ok := status["nodeInfo"].(map[string]any); ok {
			if v, ok := nodeInfo["kubeletVersion"].(string); ok {
				version = v
			}
		}

		// Check if node is ready
		ready := false
		if conditions, ok := status["conditions"].([]any); ok {
			for _, cond := range conditions {
				if condMap, ok := cond.(map[string]any); ok {
					if condMap["type"] == "Ready" && condMap["status"] == "True" {
						ready = true
						break
					}
				}
			}
		}

		// Check if node is unschedulable
		unschedulable, _ := spec["unschedulable"].(bool)

		result = append(result, map[string]any{
			"name":          name,
			"ready":         ready && !unschedulable,
			"roles":         rolesStr,
			"version":       version,
			"age":           creationTimestamp,
			"unschedulable": unschedulable,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

// GetPods returns all pods for a cluster
func (h *Handler) GetPods(w http.ResponseWriter, r *http.Request) {
	clusterID := r.PathValue("clusterId")
	if clusterID == "" {
		writeError(w, http.StatusBadRequest, "Cluster ID required")
		return
	}

	namespace := r.URL.Query().Get("namespace")

	// First, ensure we have a watcher for pods
	_ = h.clusterService.AddResourceWatcher(service.ResourceWatchRequest{
		ClusterID: clusterID,
		Group:     "",
		Version:   "v1",
		Resource:  "pods",
		Namespace: namespace,
	})

	// Load pod data
	pods, _, err := h.clusterService.LoadInitialData(clusterID, "", "v1", "pods")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Transform pods to a simpler format for the frontend
	var result []map[string]any
	for _, pod := range pods {
		metadata, _ := pod["metadata"].(map[string]any)
		spec, _ := pod["spec"].(map[string]any)
		status, _ := pod["status"].(map[string]any)

		if metadata == nil {
			continue
		}

		podNamespace, _ := metadata["namespace"].(string)
		if namespace != "" && podNamespace != namespace {
			continue
		}

		name, _ := metadata["name"].(string)
		nodeName, _ := spec["nodeName"].(string)
		phase, _ := status["phase"].(string)
		creationTimestamp, _ := metadata["creationTimestamp"].(string)

		result = append(result, map[string]any{
			"name":      name,
			"namespace": podNamespace,
			"nodeName":  nodeName,
			"phase":     phase,
			"age":       creationTimestamp,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

// LoadDefaultKubeconfig loads clusters from the default kubeconfig file
func (h *Handler) LoadDefaultKubeconfig(w http.ResponseWriter, r *http.Request) {
	kubeconfig, err := h.clusterService.LoadKubeconfigFromFile("")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Add cluster using the loaded kubeconfig
	clusterID, err := h.clusterService.AddCluster("lab-k3s", kubeconfig, "lab-k3s")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"clusterId": clusterID, "status": "loaded"})
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
