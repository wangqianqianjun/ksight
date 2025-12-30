package scheduler

import (
	"fmt"
	"strings"
	"time"
)

// Resource parsing and event handling for ClusterAggregation

// handlePodEvent processes pod add/update/delete events
func (ca *ClusterAggregation) handlePodEvent(eventType string, obj, oldObj map[string]any) {
	uid := getStringField(obj, "metadata", "uid")
	if uid == "" {
		return
	}

	switch eventType {
	case "DELETED":
		ca.removePod(uid)
	case "ADDED", "MODIFIED":
		pod := ca.parsePod(obj)
		if pod != nil {
			ca.upsertPod(pod)
		}
	}
}

// parsePod extracts PodData from unstructured object
func (ca *ClusterAggregation) parsePod(obj map[string]any) *PodData {
	metadata := getMapField(obj, "metadata")
	if metadata == nil {
		return nil
	}

	spec := getMapField(obj, "spec")
	status := getMapField(obj, "status")

	pod := &PodData{
		UID:          getStringField(metadata, "uid"),
		Namespace:    getStringField(metadata, "namespace"),
		Name:         getStringField(metadata, "name"),
		Labels:       getStringMapField(metadata, "labels"),
		Annotations:  getStringMapField(metadata, "annotations"),
		CreationTime: parseTime(getStringField(metadata, "creationTimestamp")),
	}

	if spec != nil {
		pod.NodeName = getStringField(spec, "nodeName")
		pod.SchedulerName = getStringField(spec, "schedulerName")
		if pod.SchedulerName == "" {
			pod.SchedulerName = SchedulerDefault
		}

		// Parse resourceClaims
		if claims, ok := spec["resourceClaims"].([]any); ok {
			for _, c := range claims {
				if claim, ok := c.(map[string]any); ok {
					if name := getStringField(claim, "name"); name != "" {
						pod.ResourceClaims = append(pod.ResourceClaims, name)
					}
				}
			}
		}

		// Parse resource requests from containers
		pod.GPURequest, pod.CPURequest, pod.MemoryRequest = extractPodResourceRequests(spec)
	}

	if status != nil {
		pod.Phase = getStringField(status, "phase")
	}

	return pod
}

// extractPodResourceRequests extracts aggregated resource requests from pod spec containers
func extractPodResourceRequests(spec map[string]any) (gpuRequest, cpuRequest, memoryRequest string) {
	totalGPU := 0
	var totalCPU, totalMemory int64

	containers, ok := spec["containers"].([]any)
	if !ok {
		return
	}

	for _, c := range containers {
		container, ok := c.(map[string]any)
		if !ok {
			continue
		}

		resources := getMapField(container, "resources")
		if resources == nil {
			continue
		}

		requests := getMapField(resources, "requests")
		if requests == nil {
			continue
		}

		// Extract nvidia.com/gpu
		if gpu, ok := requests["nvidia.com/gpu"]; ok {
			switch v := gpu.(type) {
			case string:
				if n := parseInt(v); n > 0 {
					totalGPU += n
				}
			case float64:
				totalGPU += int(v)
			case int64:
				totalGPU += int(v)
			}
		}

		// Extract cpu
		if cpu, ok := requests["cpu"].(string); ok {
			totalCPU += parseCPU(cpu)
		}

		// Extract memory
		if mem, ok := requests["memory"].(string); ok {
			totalMemory += parseMemory(mem)
		}
	}

	if totalGPU > 0 {
		gpuRequest = fmt.Sprintf("%d", totalGPU)
	}
	if totalCPU > 0 {
		cpuRequest = fmt.Sprintf("%dm", totalCPU)
	}
	if totalMemory > 0 {
		memoryRequest = formatBytes(totalMemory)
	}

	return
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func parseCPU(s string) int64 {
	// Parse CPU request (e.g., "100m", "1", "1.5")
	if len(s) > 0 && s[len(s)-1] == 'm' {
		var n int64
		fmt.Sscanf(s[:len(s)-1], "%d", &n)
		return n
	}
	var n float64
	fmt.Sscanf(s, "%f", &n)
	return int64(n * 1000)
}

func parseMemory(s string) int64 {
	// Parse memory request (e.g., "128Mi", "1Gi", "1024")
	multipliers := map[string]int64{
		"Ki": 1024,
		"Mi": 1024 * 1024,
		"Gi": 1024 * 1024 * 1024,
		"Ti": 1024 * 1024 * 1024 * 1024,
		"K":  1000,
		"M":  1000 * 1000,
		"G":  1000 * 1000 * 1000,
		"T":  1000 * 1000 * 1000 * 1000,
	}

	for suffix, mult := range multipliers {
		if len(s) > len(suffix) && s[len(s)-len(suffix):] == suffix {
			var n int64
			fmt.Sscanf(s[:len(s)-len(suffix)], "%d", &n)
			return n * mult
		}
	}

	var n int64
	fmt.Sscanf(s, "%d", &n)
	return n
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ci", float64(b)/float64(div), "KMGTPE"[exp])
}

// upsertPod adds or updates a pod in the cache and indexes
func (ca *ClusterAggregation) upsertPod(pod *PodData) {
	oldPod := ca.podsByUID[pod.UID]

	// Update indexes - remove old entries first
	if oldPod != nil {
		ca.removeFromIndex(ca.podsByNode, oldPod.NodeName, oldPod.UID)
		ca.removeFromIndex(ca.podsByScheduler, oldPod.SchedulerName, oldPod.UID)
	}

	// Add to cache
	ca.podsByUID[pod.UID] = pod

	// Add to indexes
	ca.addToIndex(ca.podsByNode, pod.NodeName, pod.UID)
	ca.addToIndex(ca.podsByScheduler, pod.SchedulerName, pod.UID)

	// Mark as dirty
	ca.dirtyPods[pod.UID] = struct{}{}
	if pod.NodeName != "" {
		ca.dirtyNodes[pod.NodeName] = struct{}{}
	}
}

// removePod removes a pod from cache and indexes
func (ca *ClusterAggregation) removePod(uid string) {
	pod := ca.podsByUID[uid]
	if pod == nil {
		return
	}

	// Track deletion before removing (for delta notifications)
	if pod.Phase == "Pending" && pod.NodeName == "" {
		ca.deletedPods = append(ca.deletedPods, PodRef{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		})
	}

	// Remove from indexes
	ca.removeFromIndex(ca.podsByNode, pod.NodeName, uid)
	ca.removeFromIndex(ca.podsByScheduler, pod.SchedulerName, uid)

	// Remove from cache
	delete(ca.podsByUID, uid)

	// Mark as dirty
	ca.dirtyPods[uid] = struct{}{}
	if pod.NodeName != "" {
		ca.dirtyNodes[pod.NodeName] = struct{}{}
	}
}

// handleNodeEvent processes node add/update/delete events
func (ca *ClusterAggregation) handleNodeEvent(eventType string, obj, oldObj map[string]any) {
	uid := getStringField(obj, "metadata", "uid")
	if uid == "" {
		return
	}

	switch eventType {
	case "DELETED":
		ca.removeNode(uid)
	case "ADDED", "MODIFIED":
		node := ca.parseNode(obj)
		if node != nil {
			ca.upsertNode(node)
		}
	}
}

// parseNode extracts NodeData from unstructured object
func (ca *ClusterAggregation) parseNode(obj map[string]any) *NodeData {
	metadata := getMapField(obj, "metadata")
	if metadata == nil {
		return nil
	}

	status := getMapField(obj, "status")

	node := &NodeData{
		UID:    getStringField(metadata, "uid"),
		Name:   getStringField(metadata, "name"),
		Labels: getStringMapField(metadata, "labels"),
	}

	if status != nil {
		node.Capacity = getQuantityMapField(status, "capacity")
		node.Allocatable = getQuantityMapField(status, "allocatable")
		node.Conditions = parseNodeConditions(status)
	}

	return node
}

// upsertNode adds or updates a node in the cache
func (ca *ClusterAggregation) upsertNode(node *NodeData) {
	ca.nodesByUID[node.UID] = node
	ca.dirtyNodes[node.Name] = struct{}{}
}

// removeNode removes a node from cache
func (ca *ClusterAggregation) removeNode(uid string) {
	node := ca.nodesByUID[uid]
	if node != nil {
		// Track deletion for delta notifications
		ca.deletedNodes = append(ca.deletedNodes, node.Name)
		ca.dirtyNodes[node.Name] = struct{}{}
	}
	delete(ca.nodesByUID, uid)
}

// handleEventEvent processes K8s event add/update events
func (ca *ClusterAggregation) handleEventEvent(eventType string, obj map[string]any) {
	if eventType == "DELETED" {
		return // We don't track deleted events
	}

	event := ca.parseEvent(obj)
	if event == nil {
		return
	}

	if !ca.shouldIncludeEvent(event) {
		return
	}

	// Only track scheduling-related events
	if !isSchedulingEvent(event.Reason) {
		return
	}

	// Add to event cache
	ca.eventsByUID[event.UID] = event
	ca.eventUIDOrder = append(ca.eventUIDOrder, event.UID)
	if ca.maxEvents > 0 && len(ca.eventUIDOrder) > ca.maxEvents {
		oldest := ca.eventUIDOrder[0]
		ca.eventUIDOrder = ca.eventUIDOrder[1:]
		delete(ca.eventsByUID, oldest)
	}

	// Index by pod if applicable
	if event.InvolvedObjectKind == "Pod" {
		podKey := event.InvolvedObjectNS + "/" + event.InvolvedObjectName
		events := ca.eventsByPod[podKey]

		// Keep only last 3 events per pod
		if len(events) >= 3 {
			events = events[1:]
		}
		ca.eventsByPod[podKey] = append(events, event)

		// Mark pod as dirty to update pending view
		for uid, pod := range ca.podsByUID {
			if pod.Namespace == event.InvolvedObjectNS && pod.Name == event.InvolvedObjectName {
				ca.dirtyPods[uid] = struct{}{}
				break
			}
		}
	}
}

// parseEvent extracts EventData from unstructured object
func (ca *ClusterAggregation) parseEvent(obj map[string]any) *EventData {
	metadata := getMapField(obj, "metadata")
	if metadata == nil {
		return nil
	}

	event := &EventData{
		UID:       getStringField(metadata, "uid"),
		Namespace: getStringField(metadata, "namespace"),
		Name:      getStringField(metadata, "name"),
		Reason:    getStringField(obj, "reason"),
		Message:   getStringField(obj, "note"), // events.k8s.io/v1 uses "note"
	}

	// Fallback to "message" for core/v1 events
	if event.Message == "" {
		event.Message = getStringField(obj, "message")
	}

	// Parse regarding (events.k8s.io/v1) or involvedObject (core/v1)
	if regarding := getMapField(obj, "regarding"); regarding != nil {
		event.InvolvedObjectKind = getStringField(regarding, "kind")
		event.InvolvedObjectName = getStringField(regarding, "name")
		event.InvolvedObjectNS = getStringField(regarding, "namespace")
	} else if involvedObject := getMapField(obj, "involvedObject"); involvedObject != nil {
		event.InvolvedObjectKind = getStringField(involvedObject, "kind")
		event.InvolvedObjectName = getStringField(involvedObject, "name")
		event.InvolvedObjectNS = getStringField(involvedObject, "namespace")
	}

	// Parse reporting controller
	event.ReportingController = getStringField(obj, "reportingController")
	if event.ReportingController == "" {
		if source := getMapField(obj, "source"); source != nil {
			event.ReportingController = getStringField(source, "component")
		}
	}

	// Parse timestamp
	if eventTime := getStringField(obj, "eventTime"); eventTime != "" {
		event.Timestamp = parseTime(eventTime)
	} else if lastTimestamp := getStringField(obj, "lastTimestamp"); lastTimestamp != "" {
		event.Timestamp = parseTime(lastTimestamp)
	}

	return event
}

// handleClaimEvent processes ResourceClaim events
func (ca *ClusterAggregation) handleClaimEvent(eventType string, obj, oldObj map[string]any) {
	uid := getStringField(obj, "metadata", "uid")
	if uid == "" {
		return
	}

	switch eventType {
	case "DELETED":
		ca.removeClaim(uid)
	case "ADDED", "MODIFIED":
		claim := ca.parseClaim(obj)
		if claim != nil {
			ca.upsertClaim(claim)
		}
	}
}

// parseClaim extracts ClaimData from unstructured object
func (ca *ClusterAggregation) parseClaim(obj map[string]any) *ClaimData {
	metadata := getMapField(obj, "metadata")
	if metadata == nil {
		return nil
	}

	spec := getMapField(obj, "spec")
	status := getMapField(obj, "status")

	claim := &ClaimData{
		UID:       getStringField(metadata, "uid"),
		Namespace: getStringField(metadata, "namespace"),
		Name:      getStringField(metadata, "name"),
	}

	if spec != nil {
		claim.ResourceClass = getStringField(spec, "resourceClassName")
	}

	if status != nil {
		if allocation, ok := status["allocation"].(map[string]any); ok {
			claim.Allocation = allocation

			// Extract driver from allocation
			if devices, ok := allocation["devices"].(map[string]any); ok {
				if results, ok := devices["results"].([]any); ok && len(results) > 0 {
					if first, ok := results[0].(map[string]any); ok {
						claim.Driver = getStringField(first, "driver")
					}
				}
			}
		}
	}

	// Find owning pod from ownerReferences
	if ownerRefs, ok := metadata["ownerReferences"].([]any); ok {
		for _, ref := range ownerRefs {
			if ownerRef, ok := ref.(map[string]any); ok {
				if getStringField(ownerRef, "kind") == "Pod" {
					claim.PodRef = &PodRef{
						Namespace: claim.Namespace,
						Name:      getStringField(ownerRef, "name"),
					}
					break
				}
			}
		}
	}

	return claim
}

// upsertClaim adds or updates a claim in the cache
func (ca *ClusterAggregation) upsertClaim(claim *ClaimData) {
	oldClaim := ca.claimsByUID[claim.UID]

	// Update indexes
	if oldClaim != nil && oldClaim.PodRef != nil {
		podKey := oldClaim.PodRef.Namespace + "/" + oldClaim.PodRef.Name
		ca.removeFromIndex(ca.claimsByPod, podKey, oldClaim.UID)
	}

	ca.claimsByUID[claim.UID] = claim

	if claim.PodRef != nil {
		podKey := claim.PodRef.Namespace + "/" + claim.PodRef.Name
		ca.addToIndex(ca.claimsByPod, podKey, claim.UID)
	}

	ca.dirtyClaims[claim.UID] = struct{}{}
}

// removeClaim removes a claim from cache
func (ca *ClusterAggregation) removeClaim(uid string) {
	claim := ca.claimsByUID[uid]
	if claim != nil {
		// Track deletion for delta notifications
		ca.deletedClaims = append(ca.deletedClaims, PodRef{
			Namespace: claim.Namespace,
			Name:      claim.Name,
		})

		if claim.PodRef != nil {
			podKey := claim.PodRef.Namespace + "/" + claim.PodRef.Name
			ca.removeFromIndex(ca.claimsByPod, podKey, uid)
		}
	}
	delete(ca.claimsByUID, uid)
	ca.dirtyClaims[uid] = struct{}{}
}

// handleSliceEvent processes ResourceSlice events
func (ca *ClusterAggregation) handleSliceEvent(eventType string, obj, oldObj map[string]any) {
	uid := getStringField(obj, "metadata", "uid")
	if uid == "" {
		return
	}

	switch eventType {
	case "DELETED":
		slice := ca.slicesByUID[uid]
		if slice != nil {
			// Track deletion for delta notifications
			ca.deletedSlices = append(ca.deletedSlices, slice.Name)
		}
		delete(ca.slicesByUID, uid)
		ca.dirtySlices[uid] = struct{}{}
	case "ADDED", "MODIFIED":
		slice := ca.parseSlice(obj)
		if slice != nil {
			ca.slicesByUID[slice.UID] = slice
			ca.dirtySlices[slice.UID] = struct{}{}
		}
	}
}

// parseSlice extracts SliceData from unstructured object
func (ca *ClusterAggregation) parseSlice(obj map[string]any) *SliceData {
	metadata := getMapField(obj, "metadata")
	if metadata == nil {
		return nil
	}

	spec := getMapField(obj, "spec")

	slice := &SliceData{
		UID:  getStringField(metadata, "uid"),
		Name: getStringField(metadata, "name"),
	}

	if spec != nil {
		slice.NodeName = getStringField(spec, "nodeName")
		slice.Driver = getStringField(spec, "driver")

		// Count devices
		if devices, ok := spec["devices"].([]any); ok {
			slice.Devices = len(devices)
		}
	}

	return slice
}

// handleGPUEvent processes TensorFusion GPU events
func (ca *ClusterAggregation) handleGPUEvent(eventType string, obj, oldObj map[string]any) {
	uid := getStringField(obj, "metadata", "uid")
	if uid == "" {
		return
	}

	switch eventType {
	case "DELETED":
		ca.removeGPU(uid)
	case "ADDED", "MODIFIED":
		gpu := ca.parseGPU(obj)
		if gpu != nil {
			ca.upsertGPU(gpu)
		}
	}
}

// parseGPU extracts GPUData from unstructured object
func (ca *ClusterAggregation) parseGPU(obj map[string]any) *GPUData {
	metadata := getMapField(obj, "metadata")
	if metadata == nil {
		return nil
	}

	status := getMapField(obj, "status")
	spec := getMapField(obj, "spec")

	gpu := &GPUData{
		UID:  getStringField(metadata, "uid"),
		Name: getStringField(metadata, "name"),
	}

	// Get node name from spec or labels
	if spec != nil {
		gpu.NodeName = getStringField(spec, "nodeName")
	}
	if gpu.NodeName == "" {
		labels := getStringMapField(metadata, "labels")
		if labels != nil {
			gpu.NodeName = labels["tensor-fusion.ai/node"]
		}
	}

	if status != nil {
		gpu.UsedBy = getStringField(status, "usedBy")
		gpu.Model = getStringField(status, "gpuModel") // CRD uses gpuModel, not model
		gpu.Capacity = getQuantityMapField(status, "capacity")
		gpu.Available = getQuantityMapField(status, "available")

		// Parse running apps
		if apps, ok := status["runningApps"].([]any); ok {
			for _, app := range apps {
				if appMap, ok := app.(map[string]any); ok {
					gpu.RunningApps = append(gpu.RunningApps, AppRef{
						Namespace: getStringField(appMap, "namespace"),
						Name:      getStringField(appMap, "name"),
						Count:     getIntField(appMap, "count"),
					})
				}
			}
		}
	}

	return gpu
}

// upsertGPU adds or updates a GPU in the cache
func (ca *ClusterAggregation) upsertGPU(gpu *GPUData) {
	oldGPU := ca.gpusByUID[gpu.UID]

	// Update indexes
	if oldGPU != nil {
		ca.removeFromIndex(ca.gpusByNode, oldGPU.NodeName, oldGPU.UID)
		ca.removeFromIndex(ca.gpusByUsedBy, oldGPU.UsedBy, oldGPU.UID)
	}

	ca.gpusByUID[gpu.UID] = gpu
	ca.addToIndex(ca.gpusByNode, gpu.NodeName, gpu.UID)
	ca.addToIndex(ca.gpusByUsedBy, gpu.UsedBy, gpu.UID)

	ca.dirtyGPUs[gpu.UID] = struct{}{}
	if gpu.NodeName != "" {
		ca.dirtyNodes[gpu.NodeName] = struct{}{}
	}
}

// removeGPU removes a GPU from cache
func (ca *ClusterAggregation) removeGPU(uid string) {
	gpu := ca.gpusByUID[uid]
	if gpu != nil {
		ca.removeFromIndex(ca.gpusByNode, gpu.NodeName, uid)
		ca.removeFromIndex(ca.gpusByUsedBy, gpu.UsedBy, uid)
		if gpu.NodeName != "" {
			ca.dirtyNodes[gpu.NodeName] = struct{}{}
		}
	}
	delete(ca.gpusByUID, uid)
	ca.dirtyGPUs[uid] = struct{}{}
}

// handleGPUNodeEvent processes TensorFusion GPUNode events
func (ca *ClusterAggregation) handleGPUNodeEvent(eventType string, obj, oldObj map[string]any) {
	uid := getStringField(obj, "metadata", "uid")
	if uid == "" {
		return
	}

	switch eventType {
	case "DELETED":
		gpuNode := ca.gpuNodesByUID[uid]
		if gpuNode != nil {
			ca.dirtyNodes[gpuNode.NodeName] = struct{}{}
		}
		delete(ca.gpuNodesByUID, uid)
	case "ADDED", "MODIFIED":
		gpuNode := ca.parseGPUNode(obj)
		if gpuNode != nil {
			ca.gpuNodesByUID[gpuNode.UID] = gpuNode
			ca.dirtyNodes[gpuNode.NodeName] = struct{}{}
		}
	}
}

// parseGPUNode extracts GPUNodeData from unstructured object
func (ca *ClusterAggregation) parseGPUNode(obj map[string]any) *GPUNodeData {
	metadata := getMapField(obj, "metadata")
	if metadata == nil {
		return nil
	}

	status := getMapField(obj, "status")
	spec := getMapField(obj, "spec")

	gpuNode := &GPUNodeData{
		UID:  getStringField(metadata, "uid"),
		Name: getStringField(metadata, "name"),
	}

	// Node name from spec or name
	if spec != nil {
		gpuNode.NodeName = getStringField(spec, "nodeName")
	}
	if gpuNode.NodeName == "" {
		gpuNode.NodeName = gpuNode.Name
	}

	// Pool from labels
	labels := getStringMapField(metadata, "labels")
	if labels != nil {
		gpuNode.Pool = labels["tensor-fusion.ai/gpupool"]
	}

	if status != nil {
		gpuNode.Phase = getStringField(status, "phase")
		gpuNode.ManagedGPUs = int32(getIntField(status, "managedGPUs"))
	}

	return gpuNode
}

// handleGPUPoolEvent processes TensorFusion GPUPool events
func (ca *ClusterAggregation) handleGPUPoolEvent(eventType string, obj, oldObj map[string]any) {
	uid := getStringField(obj, "metadata", "uid")
	if uid == "" {
		return
	}

	switch eventType {
	case "DELETED":
		delete(ca.gpuPoolsByUID, uid)
	case "ADDED", "MODIFIED":
		pool := ca.parseGPUPool(obj)
		if pool != nil {
			ca.gpuPoolsByUID[pool.UID] = pool
		}
	}
}

// parseGPUPool extracts GPUPoolData from unstructured object
func (ca *ClusterAggregation) parseGPUPool(obj map[string]any) *GPUPoolData {
	metadata := getMapField(obj, "metadata")
	if metadata == nil {
		return nil
	}

	spec := getMapField(obj, "spec")
	status := getMapField(obj, "status")

	pool := &GPUPoolData{
		UID:  getStringField(metadata, "uid"),
		Name: getStringField(metadata, "name"),
	}

	if spec != nil {
		// Check oversubscription
		if capacityConfig := getMapField(spec, "capacityConfig"); capacityConfig != nil {
			if oversubscription := getMapField(capacityConfig, "oversubscription"); oversubscription != nil {
				pool.Oversubscription = true
			}
		}

		// Check DRA enabled (CRD field is "enable", not "enabled")
		if draConfig := getMapField(spec, "draConfig"); draConfig != nil {
			pool.DRAEnabled = getBoolField(draConfig, "enable")
		}
	}

	if status != nil {
		pool.Phase = getStringField(status, "phase")
		pool.TotalGPUs = int32(getIntField(status, "totalGPUs"))
		pool.AvailableVRAM = getInt64Field(status, "availableVRAM")
		pool.AvailableTFlops = getFloatField(status, "availableTFlops")
	}

	return pool
}

// handleDeploymentEvent processes Deployment events for progressive migration detection
func (ca *ClusterAggregation) handleDeploymentEvent(eventType string, obj, oldObj map[string]any) {
	uid := getStringField(obj, "metadata", "uid")
	if uid == "" {
		return
	}

	switch eventType {
	case "DELETED":
		delete(ca.deploymentsByUID, uid)
		ca.dirtyMigration = true
	case "ADDED", "MODIFIED":
		deployment := ca.parseDeployment(obj)
		if deployment != nil {
			ca.deploymentsByUID[deployment.UID] = deployment
			ca.dirtyMigration = true
		}
	}
}

// parseDeployment extracts DeploymentData from unstructured object
func (ca *ClusterAggregation) parseDeployment(obj map[string]any) *DeploymentData {
	metadata := getMapField(obj, "metadata")
	if metadata == nil {
		return nil
	}

	spec := getMapField(obj, "spec")
	if spec == nil {
		return nil
	}

	template := getMapField(spec, "template")
	podSpec := getMapField(template, "spec")

	deployment := &DeploymentData{
		UID:       getStringField(metadata, "uid"),
		Namespace: getStringField(metadata, "namespace"),
		Name:      getStringField(metadata, "name"),
		Labels:    getStringMapField(metadata, "labels"),
	}

	if podSpec == nil {
		return deployment
	}

	containers, ok := podSpec["containers"].([]any)
	if !ok {
		return deployment
	}

	for _, c := range containers {
		container, ok := c.(map[string]any)
		if !ok {
			continue
		}
		env, ok := container["env"].([]any)
		if !ok {
			continue
		}
		for _, e := range env {
			envVar, ok := e.(map[string]any)
			if !ok {
				continue
			}
			if getStringField(envVar, "name") == "NVIDIA_OPERATOR_PROGRESSIVE_MIGRATION" {
				val := getStringField(envVar, "value")
				parsed := parseBoolString(val)
				deployment.ProgressiveMigration = parsed
				return deployment
			}
		}
	}

	return deployment
}

// Index helper methods

func (ca *ClusterAggregation) addToIndex(index map[string]map[string]struct{}, key, value string) {
	if key == "" {
		return
	}
	if index[key] == nil {
		index[key] = make(map[string]struct{})
	}
	index[key][value] = struct{}{}
}

func (ca *ClusterAggregation) removeFromIndex(index map[string]map[string]struct{}, key, value string) {
	if key == "" {
		return
	}
	if index[key] != nil {
		delete(index[key], value)
		if len(index[key]) == 0 {
			delete(index, key)
		}
	}
}

// Helper functions for parsing unstructured objects

func getMapField(obj map[string]any, keys ...string) map[string]any {
	current := obj
	for _, key := range keys {
		if val, ok := current[key].(map[string]any); ok {
			current = val
		} else {
			return nil
		}
	}
	return current
}

func getStringField(obj map[string]any, keys ...string) string {
	if len(keys) == 0 {
		return ""
	}

	current := obj
	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key].(string); ok {
				return val
			}
			return ""
		}
		if next, ok := current[key].(map[string]any); ok {
			current = next
		} else {
			return ""
		}
	}
	return ""
}

func getStringMapField(obj map[string]any, key string) map[string]string {
	if val, ok := obj[key].(map[string]any); ok {
		result := make(map[string]string)
		for k, v := range val {
			if str, ok := v.(string); ok {
				result[k] = str
			}
		}
		return result
	}
	return nil
}

func getQuantityMapField(obj map[string]any, key string) map[string]string {
	if val, ok := obj[key].(map[string]any); ok {
		result := make(map[string]string)
		for k, v := range val {
			switch t := v.(type) {
			case string:
				result[k] = t
			case float64:
				result[k] = fmt.Sprintf("%.0f", t)
			case int64:
				result[k] = fmt.Sprintf("%d", t)
			}
		}
		return result
	}
	return nil
}

func getIntField(obj map[string]any, key string) int {
	if val, ok := obj[key].(float64); ok {
		return int(val)
	}
	if val, ok := obj[key].(int); ok {
		return val
	}
	if val, ok := obj[key].(int64); ok {
		return int(val)
	}
	return 0
}

func getInt64Field(obj map[string]any, key string) int64 {
	if val, ok := obj[key].(float64); ok {
		return int64(val)
	}
	if val, ok := obj[key].(int64); ok {
		return val
	}
	if val, ok := obj[key].(int); ok {
		return int64(val)
	}
	return 0
}

func getFloatField(obj map[string]any, key string) float64 {
	if val, ok := obj[key].(float64); ok {
		return val
	}
	return 0
}

func getBoolField(obj map[string]any, key string) bool {
	if val, ok := obj[key].(bool); ok {
		return val
	}
	return false
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseNodeConditions(status map[string]any) map[string]string {
	conditions := make(map[string]string)
	if conds, ok := status["conditions"].([]any); ok {
		for _, c := range conds {
			if cond, ok := c.(map[string]any); ok {
				condType := getStringField(cond, "type")
				condStatus := getStringField(cond, "status")
				if condType != "" {
					conditions[condType] = condStatus
				}
			}
		}
	}
	return conditions
}

func isSchedulingEvent(reason string) bool {
	switch reason {
	case ReasonFailedScheduling, ReasonScheduled,
		ReasonGPUQuotaOrCapacityNotEnough, ReasonScheduleWithNativeGPU:
		return true
	}
	return false
}

func parseBoolString(value string) *bool {
	if value == "" {
		return nil
	}
	switch strings.ToLower(value) {
	case "true", "1", "yes":
		v := true
		return &v
	case "false", "0", "no":
		v := false
		return &v
	default:
		return nil
	}
}
