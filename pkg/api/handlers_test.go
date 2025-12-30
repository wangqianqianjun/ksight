package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ksight/pkg/service"
	"ksight/pkg/service/scheduler"

	"github.com/gorilla/websocket"
)

type fakeClusterService struct {
	clusters              map[string]service.ClusterInfo
	addClusterID          string
	addClusterErr         error
	removeClusterErr      error
	loadInitialErr        error
	loadInitialData       map[string][]map[string]any
	loadKubeconfigErr     error
	kubeconfig            string
	addWatcherRequests    []service.ResourceWatchRequest
	removeWatcherRequests []service.ResourceWatchRequest
}

func (f *fakeClusterService) AddCluster(name, kubeconfig, context string) (string, error) {
	if f.addClusterErr != nil {
		return "", f.addClusterErr
	}
	if f.addClusterID == "" {
		return "cluster-1", nil
	}
	return f.addClusterID, nil
}

func (f *fakeClusterService) RemoveCluster(clusterID string) error {
	return f.removeClusterErr
}

func (f *fakeClusterService) GetClusters() map[string]service.ClusterInfo {
	if f.clusters == nil {
		return map[string]service.ClusterInfo{}
	}
	return f.clusters
}

func (f *fakeClusterService) AddResourceWatcher(request service.ResourceWatchRequest) error {
	f.addWatcherRequests = append(f.addWatcherRequests, request)
	return nil
}

func (f *fakeClusterService) RemoveResourceWatcher(request service.ResourceWatchRequest) error {
	f.removeWatcherRequests = append(f.removeWatcherRequests, request)
	return nil
}

func (f *fakeClusterService) LoadInitialData(clusterID string, group, version, resource string) ([]map[string]any, string, error) {
	if f.loadInitialErr != nil {
		return nil, "", f.loadInitialErr
	}
	if f.loadInitialData == nil {
		return nil, "", nil
	}
	key := group + "/" + version + "/" + resource
	return f.loadInitialData[key], "", nil
}

func (f *fakeClusterService) LoadKubeconfigFromFile(filePath string) (string, error) {
	if f.loadKubeconfigErr != nil {
		return "", f.loadKubeconfigErr
	}
	if f.kubeconfig != "" {
		return f.kubeconfig, nil
	}
	return "kubeconfig", nil
}

type fakeSchedulerAggregator struct {
	startErr    error
	stopErr     error
	snapshotErr error
	startReq    *scheduler.SchedulerAggregationRequest
	stopID      string
	snapshot    *scheduler.SchedulerSnapshot
}

func (f *fakeSchedulerAggregator) StartAggregation(request scheduler.SchedulerAggregationRequest) error {
	f.startReq = &request
	return f.startErr
}

func (f *fakeSchedulerAggregator) StopAggregation(clusterID string) error {
	f.stopID = clusterID
	return f.stopErr
}

func (f *fakeSchedulerAggregator) GetSnapshot(clusterID string) (*scheduler.SchedulerSnapshot, error) {
	if f.snapshotErr != nil {
		return nil, f.snapshotErr
	}
	if f.snapshot != nil {
		return f.snapshot, nil
	}
	return &scheduler.SchedulerSnapshot{ClusterID: clusterID}, nil
}

func decodeJSON[T any](t *testing.T, rr *httptest.ResponseRecorder, target *T) {
	t.Helper()
	if err := json.NewDecoder(rr.Body).Decode(target); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestGetClusters(t *testing.T) {
	fakeCS := &fakeClusterService{
		clusters: map[string]service.ClusterInfo{
			"c1": {ID: "c1", Name: "cluster-1"},
		},
	}
	handler := NewHandler(fakeCS, &fakeSchedulerAggregator{}, NewWebSocketHub())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/clusters", nil)
	handler.GetClusters(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]service.ClusterInfo
	decodeJSON(t, rr, &body)
	if body["c1"].Name != "cluster-1" {
		t.Fatalf("expected cluster name, got %+v", body["c1"])
	}
}

func TestAddCluster(t *testing.T) {
	fakeCS := &fakeClusterService{addClusterID: "cluster-9"}
	handler := NewHandler(fakeCS, &fakeSchedulerAggregator{}, NewWebSocketHub())

	payload := `{"name":"demo","kubeconfig":"cfg","context":"ctx"}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/clusters", strings.NewReader(payload))
	handler.AddCluster(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]string
	decodeJSON(t, rr, &body)
	if body["clusterId"] != "cluster-9" {
		t.Fatalf("expected clusterId cluster-9, got %v", body["clusterId"])
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/clusters", strings.NewReader("{bad"))
	handler.AddCluster(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestRemoveCluster(t *testing.T) {
	fakeCS := &fakeClusterService{}
	handler := NewHandler(fakeCS, &fakeSchedulerAggregator{}, NewWebSocketHub())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/clusters/", nil)
	handler.RemoveCluster(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/clusters/c1", nil)
	req.SetPathValue("id", "c1")
	handler.RemoveCluster(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	fakeCS.removeClusterErr = errors.New("boom")
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/clusters/c1", nil)
	req.SetPathValue("id", "c1")
	handler.RemoveCluster(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

func TestStartSchedulerAggregation(t *testing.T) {
	fakeAgg := &fakeSchedulerAggregator{}
	handler := NewHandler(&fakeClusterService{}, fakeAgg, NewWebSocketHub())

	payload := `{"clusterId":"c1","include":{"pods":true}}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/scheduler/start", strings.NewReader(payload))
	handler.StartSchedulerAggregation(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if fakeAgg.startReq == nil || fakeAgg.startReq.ClusterID != "c1" {
		t.Fatalf("expected StartAggregation to be called with clusterId c1")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/scheduler/start", strings.NewReader("{bad"))
	handler.StartSchedulerAggregation(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	fakeAgg.startErr = errors.New("boom")
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/scheduler/start", bytes.NewBufferString(payload))
	handler.StartSchedulerAggregation(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

func TestStopSchedulerAggregation(t *testing.T) {
	fakeAgg := &fakeSchedulerAggregator{}
	handler := NewHandler(&fakeClusterService{}, fakeAgg, NewWebSocketHub())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/scheduler/stop", strings.NewReader(`{"clusterId":"c1"}`))
	handler.StopSchedulerAggregation(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if fakeAgg.stopID != "c1" {
		t.Fatalf("expected StopAggregation to be called with c1")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/scheduler/stop", strings.NewReader("{bad"))
	handler.StopSchedulerAggregation(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	fakeAgg.stopErr = errors.New("boom")
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/scheduler/stop", strings.NewReader(`{"clusterId":"c1"}`))
	handler.StopSchedulerAggregation(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

func TestGetSchedulerSnapshot(t *testing.T) {
	fakeAgg := &fakeSchedulerAggregator{
		snapshot: &scheduler.SchedulerSnapshot{ClusterID: "c1"},
	}
	handler := NewHandler(&fakeClusterService{}, fakeAgg, NewWebSocketHub())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/scheduler/snapshot/", nil)
	handler.GetSchedulerSnapshot(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/scheduler/snapshot/c1", nil)
	req.SetPathValue("clusterId", "c1")
	handler.GetSchedulerSnapshot(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var snapshot scheduler.SchedulerSnapshot
	decodeJSON(t, rr, &snapshot)
	if snapshot.ClusterID != "c1" {
		t.Fatalf("expected clusterId c1, got %s", snapshot.ClusterID)
	}

	fakeAgg.snapshotErr = errors.New("boom")
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/scheduler/snapshot/c1", nil)
	req.SetPathValue("clusterId", "c1")
	handler.GetSchedulerSnapshot(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

func TestGetNodes(t *testing.T) {
	fakeCS := &fakeClusterService{}
	handler := NewHandler(fakeCS, &fakeSchedulerAggregator{}, NewWebSocketHub())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/clusters//nodes", nil)
	handler.GetNodes(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	fakeCS.loadInitialErr = errors.New("boom")
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/clusters/c1/nodes", nil)
	req.SetPathValue("clusterId", "c1")
	handler.GetNodes(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	fakeCS.loadInitialErr = nil
	fakeCS.loadInitialData = map[string][]map[string]any{
		"/v1/nodes": {
			{
				"metadata": map[string]any{
					"name":              "node-1",
					"labels":            map[string]any{"node-role.kubernetes.io/master": "", "kubernetes.io/hostname": "node-1"},
					"creationTimestamp": time.Now().Format(time.RFC3339),
				},
				"spec": map[string]any{
					"unschedulable": false,
				},
				"status": map[string]any{
					"nodeInfo": map[string]any{"kubeletVersion": "v1.30.0"},
					"conditions": []any{
						map[string]any{"type": "Ready", "status": "True"},
					},
				},
			},
		},
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/clusters/c1/nodes", nil)
	req.SetPathValue("clusterId", "c1")
	handler.GetNodes(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var nodes []map[string]any
	decodeJSON(t, rr, &nodes)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0]["name"] != "node-1" {
		t.Fatalf("expected node name node-1, got %v", nodes[0]["name"])
	}
}

func TestGetPods(t *testing.T) {
	fakeCS := &fakeClusterService{}
	handler := NewHandler(fakeCS, &fakeSchedulerAggregator{}, NewWebSocketHub())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/clusters//pods", nil)
	handler.GetPods(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	fakeCS.loadInitialErr = errors.New("boom")
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/clusters/c1/pods", nil)
	req.SetPathValue("clusterId", "c1")
	handler.GetPods(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	fakeCS.loadInitialErr = nil
	fakeCS.loadInitialData = map[string][]map[string]any{
		"/v1/pods": {
			{
				"metadata": map[string]any{
					"name":              "pod-1",
					"namespace":         "ns-1",
					"creationTimestamp": time.Now().Format(time.RFC3339),
				},
				"spec": map[string]any{
					"nodeName": "node-1",
				},
				"status": map[string]any{
					"phase": "Running",
				},
			},
			{
				"metadata": map[string]any{
					"name":              "pod-2",
					"namespace":         "ns-2",
					"creationTimestamp": time.Now().Format(time.RFC3339),
				},
				"spec": map[string]any{},
				"status": map[string]any{
					"phase": "Pending",
				},
			},
		},
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/clusters/c1/pods?namespace=ns-1", nil)
	req.SetPathValue("clusterId", "c1")
	handler.GetPods(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var pods []map[string]any
	decodeJSON(t, rr, &pods)
	if len(pods) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(pods))
	}
	if pods[0]["name"] != "pod-1" {
		t.Fatalf("expected pod name pod-1, got %v", pods[0]["name"])
	}
}

func TestLoadDefaultKubeconfig(t *testing.T) {
	fakeCS := &fakeClusterService{}
	handler := NewHandler(fakeCS, &fakeSchedulerAggregator{}, NewWebSocketHub())

	fakeCS.loadKubeconfigErr = errors.New("boom")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/clusters/load-default", nil)
	handler.LoadDefaultKubeconfig(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	fakeCS.loadKubeconfigErr = nil
	fakeCS.addClusterErr = errors.New("boom")
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/clusters/load-default", nil)
	handler.LoadDefaultKubeconfig(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	fakeCS.addClusterErr = nil
	fakeCS.addClusterID = "cluster-9"
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/clusters/load-default", nil)
	handler.LoadDefaultKubeconfig(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestHandleWebSocket(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping websocket test; listener unavailable: %v", err)
	}
	listener.Close()

	hub := NewWebSocketHub()
	go hub.Run()

	handler := NewHandler(&fakeClusterService{}, &fakeSchedulerAggregator{}, hub)
	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	conn.Close()
}
