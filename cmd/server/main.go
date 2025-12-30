package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ksight/pkg/api"
	"ksight/pkg/informer"
	"ksight/pkg/service"
	"ksight/pkg/service/scheduler"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create cluster service
	clusterService := service.NewClusterService(ctx)

	// Create WebSocket hub for real-time events
	wsHub := api.NewWebSocketHub()
	go wsHub.Run()

	// Create event emitter that broadcasts to WebSocket clients
	wsEmitter := &api.WebSocketEventEmitter{Hub: wsHub}

	// Create scheduler aggregator
	schedulerAggregator := scheduler.NewAggregator(
		wsEmitter,
		&clusterClientAdapter{cs: clusterService},
	)

	// Wire informer events to both scheduler aggregator and WebSocket
	clusterService.RegisterResourceHandler(func(event informer.Event) {
		var obj map[string]any
		var oldObj map[string]any

		if event.Object != nil {
			obj = event.Object.UnstructuredContent()
		}
		if event.OldObject != nil {
			oldObj = event.OldObject.UnstructuredContent()
		}

		schedulerAggregator.HandleEvent(
			event.ClusterID,
			event.GVR,
			event.Type,
			obj,
			oldObj,
		)

		// Also emit raw resource events to WebSocket
		wsEmitter.Emit("resource:event", event)
	})

	// Create API handler
	apiHandler := api.NewHandler(clusterService, schedulerAggregator, wsHub)

	// Auto-load default kubeconfig on startup
	go func() {
		kubeconfig, err := clusterService.LoadKubeconfigFromFile("")
		if err != nil {
			log.Printf("Warning: Failed to load default kubeconfig: %v", err)
			return
		}
		clusterID, err := clusterService.AddCluster("lab-k3s", kubeconfig, "lab-k3s")
		if err != nil {
			log.Printf("Warning: Failed to add default cluster: %v", err)
			return
		}
		log.Printf("Auto-loaded cluster: %s", clusterID)
	}()

	// Setup routes
	mux := http.NewServeMux()

	// Cluster APIs
	mux.HandleFunc("GET /api/clusters", apiHandler.GetClusters)
	mux.HandleFunc("POST /api/clusters", apiHandler.AddCluster)
	mux.HandleFunc("DELETE /api/clusters/{id}", apiHandler.RemoveCluster)
	mux.HandleFunc("POST /api/clusters/load-default", apiHandler.LoadDefaultKubeconfig)
	mux.HandleFunc("GET /api/clusters/{clusterId}/pods", apiHandler.GetPods)
	mux.HandleFunc("GET /api/clusters/{clusterId}/nodes", apiHandler.GetNodes)

	// Scheduler APIs
	mux.HandleFunc("POST /api/scheduler/start", apiHandler.StartSchedulerAggregation)
	mux.HandleFunc("POST /api/scheduler/stop", apiHandler.StopSchedulerAggregation)
	mux.HandleFunc("GET /api/scheduler/snapshot/{clusterId}", apiHandler.GetSchedulerSnapshot)

	// WebSocket for real-time events
	mux.HandleFunc("GET /api/ws", apiHandler.HandleWebSocket)

	// CORS middleware
	handler := corsMiddleware(mux)

	// Create server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("API server starting on http://localhost:8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	clusterService.Shutdown()
	log.Println("Server stopped")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// clusterClientAdapter adapts ClusterService to scheduler.ClusterClientProvider
type clusterClientAdapter struct {
	cs *service.ClusterService
}

func (a *clusterClientAdapter) AddResourceWatcher(clusterID string, gvr schema.GroupVersionResource, namespace string) error {
	return a.cs.AddResourceWatcher(service.ResourceWatchRequest{
		ClusterID: clusterID,
		Group:     gvr.Group,
		Version:   gvr.Version,
		Resource:  gvr.Resource,
		Namespace: namespace,
	})
}

func (a *clusterClientAdapter) RemoveResourceWatcher(clusterID string, gvr schema.GroupVersionResource) error {
	return a.cs.RemoveResourceWatcher(service.ResourceWatchRequest{
		ClusterID: clusterID,
		Group:     gvr.Group,
		Version:   gvr.Version,
		Resource:  gvr.Resource,
	})
}
