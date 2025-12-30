package scheduler

import (
	"fmt"
	"sort"
	"time"
)

// View generation methods for ClusterAggregation

// generateNodeView generates the full node view snapshot
func (ca *ClusterAggregation) generateNodeView() NodeViewSnapshot {
	var nodes []NodeRow
	summary := NodeSummary{
		GPUsByUsedBy: make(map[string]int),
		GPUsByModel:  make(map[string]int),
	}

	// Build a map of node name -> GPUNode for quick lookup
	gpuNodeByNodeName := make(map[string]*GPUNodeData)
	for _, gpuNode := range ca.gpuNodesByUID {
		gpuNodeByNodeName[gpuNode.NodeName] = gpuNode
	}

	// Build node rows
	for _, node := range ca.nodesByUID {
		row := ca.buildNodeRow(node, gpuNodeByNodeName)
		nodes = append(nodes, row)

		// Update summary
		summary.TotalNodes++
		if node.Conditions["Ready"] == "True" {
			summary.ReadyNodes++
		}
		summary.TotalGPUs += row.GPU.Total
		for usedBy, count := range row.GPU.UsedBy {
			summary.GPUsByUsedBy[usedBy] += count
		}
	}

	// Calculate TFlops and VRAM from GPUPools
	for _, pool := range ca.gpuPoolsByUID {
		summary.AvailableTFlops += pool.AvailableTFlops
		summary.AvailableVRAM += pool.AvailableVRAM
	}

	// Sort nodes by name
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	return NodeViewSnapshot{
		Nodes:   nodes,
		Summary: summary,
	}
}

// buildNodeRow builds a NodeRow from node data
func (ca *ClusterAggregation) buildNodeRow(node *NodeData, gpuNodeByNodeName map[string]*GPUNodeData) NodeRow {
	row := NodeRow{
		Name:        node.Name,
		Labels:      node.Labels,
		Capacity:    node.Capacity,
		Allocatable: node.Allocatable,
	}

	// Extract zone from labels
	if node.Labels != nil {
		if zone, ok := node.Labels["topology.kubernetes.io/zone"]; ok {
			row.Zone = zone
		} else if zone, ok := node.Labels["failure-domain.beta.kubernetes.io/zone"]; ok {
			row.Zone = zone
		}
	}

	// Add TensorFusion info
	if gpuNode, ok := gpuNodeByNodeName[node.Name]; ok {
		row.TensorFusion = NodeTensorFusionInfo{
			Pool:        gpuNode.Pool,
			Phase:       gpuNode.Phase,
			ManagedGPUs: gpuNode.ManagedGPUs,
		}
	}

	// Add GPU info
	row.GPU = NodeGPUInfo{
		UsedBy: make(map[string]int),
	}
	gpuUIDs := ca.gpusByNode[node.Name]
	for uid := range gpuUIDs {
		gpu := ca.gpusByUID[uid]
		if gpu == nil {
			continue
		}

		row.GPU.Total++

		usedBy := gpu.UsedBy
		if usedBy == "" {
			usedBy = UsedByUnknown
		}
		row.GPU.UsedBy[usedBy]++

		row.GPU.Devices = append(row.GPU.Devices, GPUDeviceRow{
			Name:        gpu.Name,
			Model:       gpu.Model,
			UsedBy:      gpu.UsedBy,
			Capacity:    gpu.Capacity,
			Available:   gpu.Available,
			RunningApps: gpu.RunningApps,
		})
	}

	// Also check native nvidia.com/gpu from node allocatable
	if row.Allocatable != nil {
		if nvidiaGPU, ok := row.Allocatable["nvidia.com/gpu"]; ok && nvidiaGPU != "0" {
			// If we don't have TensorFusion GPUs on this node, count native GPUs
			if row.GPU.Total == 0 {
				// This is a fallback - actual count parsing would need more work
				row.GPU.UsedBy[UsedByNvidiaDevicePlugin] = 1
			}
		}
	}

	// Add pods on this node (if minPayload is false)
	if !ca.minPayload {
		podUIDs := ca.podsByNode[node.Name]
		for uid := range podUIDs {
			pod := ca.podsByUID[uid]
			if pod != nil {
				row.Pods = append(row.Pods, PodRef{
					Namespace: pod.Namespace,
					Name:      pod.Name,
				})
			}
		}
	}

	return row
}

// generatePendingView generates the full pending pod view snapshot
func (ca *ClusterAggregation) generatePendingView() PendingViewSnapshot {
	// Temporary struct for sorting by creation time
	type pendingPodWithTime struct {
		row          PendingPodRow
		creationTime time.Time
	}

	var podsWithTime []pendingPodWithTime
	byScheduler := make(map[string]int)
	reasonCounts := make(map[string]int)

	for _, pod := range ca.podsByUID {
		// Only include pending pods without a node assigned
		if pod.Phase != "Pending" || pod.NodeName != "" {
			continue
		}

		row := ca.buildPendingPodRow(pod)
		podsWithTime = append(podsWithTime, pendingPodWithTime{
			row:          row,
			creationTime: pod.CreationTime,
		})

		byScheduler[row.Scheduler]++
		if row.Reason != "" {
			reasonCounts[row.Reason]++
		}
	}

	// Sort by creation time (oldest first - smaller time is earlier)
	sort.Slice(podsWithTime, func(i, j int) bool {
		return podsWithTime[i].creationTime.Before(podsWithTime[j].creationTime)
	})

	// Extract just the rows
	pods := make([]PendingPodRow, len(podsWithTime))
	for i, p := range podsWithTime {
		pods[i] = p.row
	}

	// Convert reason counts to buckets
	var reasons []ReasonBucket
	for reason, count := range reasonCounts {
		reasons = append(reasons, ReasonBucket{
			Reason: reason,
			Count:  count,
		})
	}
	sort.Slice(reasons, func(i, j int) bool {
		return reasons[i].Count > reasons[j].Count
	})

	return PendingViewSnapshot{
		Pods:        pods,
		ByScheduler: byScheduler,
		Reasons:     reasons,
	}
}

// buildPendingPodRow builds a PendingPodRow from pod data
func (ca *ClusterAggregation) buildPendingPodRow(pod *PodData) PendingPodRow {
	row := PendingPodRow{
		Namespace: pod.Namespace,
		Name:      pod.Name,
		Scheduler: pod.SchedulerName,
	}

	// Extract GPU request from annotations (TensorFusion format)
	if pod.Annotations != nil {
		if gpuCount, ok := pod.Annotations[TFAnnotationGPUCount]; ok {
			row.GPURequest = gpuCount
		} else if tflops, ok := pod.Annotations[TFAnnotationTFlopsRequest]; ok {
			row.GPURequest = tflops + " TFlops"
		} else if vram, ok := pod.Annotations[TFAnnotationVRAMRequest]; ok {
			row.GPURequest = vram + " VRAM"
		}

		if pool, ok := pod.Annotations[TFAnnotationGPUPool]; ok {
			row.Pool = pool
		}
	}

	// Fallback to nvidia.com/gpu from pod resource requests if TF annotations not present
	if row.GPURequest == "" && pod.GPURequest != "" {
		row.GPURequest = pod.GPURequest + " GPU"
	}

	// Get latest failure reason from events
	podKey := pod.Namespace + "/" + pod.Name
	if events := ca.eventsByPod[podKey]; len(events) > 0 {
		// Get most recent event
		latest := events[len(events)-1]
		if latest.Reason == ReasonFailedScheduling || latest.Reason == ReasonGPUQuotaOrCapacityNotEnough {
			row.Reason = latest.Message
			if len(row.Reason) > 200 {
				row.Reason = row.Reason[:200] + "..."
			}
		}
	}

	// Calculate pending duration
	if !pod.CreationTime.IsZero() {
		duration := time.Since(pod.CreationTime)
		row.Since = formatDuration(duration)
	}

	return row
}

// generateDRAView generates the full DRA view snapshot
func (ca *ClusterAggregation) generateDRAView() DRASnapshot {
	var claims []ClaimRow
	var slices []SliceRow
	stats := make(map[string]int)

	// Process claims
	for _, claim := range ca.claimsByUID {
		row := ClaimRow{
			Namespace:     claim.Namespace,
			Name:          claim.Name,
			ResourceClass: claim.ResourceClass,
			Driver:        claim.Driver,
			AllocationRaw: claim.Allocation,
			PodRef:        claim.PodRef,
		}
		claims = append(claims, row)

		// Update stats
		if claim.Allocation != nil {
			stats["allocated"]++
		} else {
			stats["pending"]++
		}
	}

	// Process slices
	for _, slice := range ca.slicesByUID {
		row := SliceRow{
			Name:     slice.Name,
			NodeName: slice.NodeName,
			Driver:   slice.Driver,
			Devices:  slice.Devices,
		}
		slices = append(slices, row)
		stats["totalDevices"] += slice.Devices
	}

	// Sort by name
	sort.Slice(claims, func(i, j int) bool {
		if claims[i].Namespace != claims[j].Namespace {
			return claims[i].Namespace < claims[j].Namespace
		}
		return claims[i].Name < claims[j].Name
	})
	sort.Slice(slices, func(i, j int) bool {
		return slices[i].Name < slices[j].Name
	})

	return DRASnapshot{
		Claims: claims,
		Slices: slices,
		Stats:  stats,
	}
}

// generateMigrationView generates the full migration/coexistence view snapshot
func (ca *ClusterAggregation) generateMigrationView() MigrationSnapshot {
	usedByStats := make(map[string]int)
	var conflicts []ConflictRow

	// Count GPUs by usedBy
	for _, gpu := range ca.gpusByUID {
		usedBy := gpu.UsedBy
		if usedBy == "" {
			usedBy = UsedByUnknown
		}
		usedByStats[usedBy]++
	}

	// Detect conflicts: pods using nvidia.com/gpu but on nodes managed by TensorFusion
	for _, pod := range ca.podsByUID {
		if pod.NodeName == "" {
			continue
		}

		// Check if pod requests native nvidia.com/gpu
		// This would need to be extracted from pod spec containers
		// For now, we detect based on scheduler mismatch
		if pod.SchedulerName == SchedulerTensorFusion {
			// Check if any annotation indicates native GPU usage
			if pod.Annotations != nil {
				if _, hasTF := pod.Annotations[TFAnnotationGPUPool]; !hasTF {
					// Pod scheduled by TF scheduler but no TF annotations - potential conflict
					// This is a simplified check
				}
			}
		}

		// Check for actual GPU usedBy mismatch
		gpuUIDs := ca.gpusByNode[pod.NodeName]
		for uid := range gpuUIDs {
			gpu := ca.gpusByUID[uid]
			if gpu == nil {
				continue
			}

			// Check if pod is using this GPU but usedBy doesn't match
			for _, app := range gpu.RunningApps {
				if app.Namespace == pod.Namespace && app.Name == pod.Name {
					// Pod is running on this GPU
					expectedUsedBy := UsedByTensorFusion
					if pod.SchedulerName != SchedulerTensorFusion {
						expectedUsedBy = UsedByNvidiaDevicePlugin
					}

					if gpu.UsedBy != "" && gpu.UsedBy != expectedUsedBy {
						conflicts = append(conflicts, ConflictRow{
							Node: pod.NodeName,
							Pod: PodRef{
								Namespace: pod.Namespace,
								Name:      pod.Name,
							},
							UsedBy:       gpu.UsedBy,
							ResourceType: determineResourceType(pod),
						})
					}
				}
			}
		}
	}

	// Check progressive migration mode from GPUPools
	progressiveMigration := false
	for _, pool := range ca.gpuPoolsByUID {
		// If any pool has oversubscription enabled, consider it progressive migration mode
		if pool.Oversubscription {
			progressiveMigration = true
			break
		}
	}

	return MigrationSnapshot{
		UsedByStats:          usedByStats,
		Conflicts:            conflicts,
		ProgressiveMigration: progressiveMigration,
	}
}

// determineResourceType determines what GPU resource type a pod is using
func determineResourceType(pod *PodData) string {
	if pod.Annotations != nil {
		if _, ok := pod.Annotations[TFAnnotationDRAEnabled]; ok {
			return "dra"
		}
		if _, ok := pod.Annotations[TFAnnotationGPUPool]; ok {
			return "tensor-fusion"
		}
	}
	return "nvidia.com/gpu"
}

// Delta generation methods

// generateNodeViewDelta generates delta for node view
func (ca *ClusterAggregation) generateNodeViewDelta() NodeViewDelta {
	delta := NodeViewDelta{}

	// Include deleted nodes in removal list
	delta.Remove = append(delta.Remove, ca.deletedNodes...)

	// Build gpuNodeByNodeName map
	gpuNodeByNodeName := make(map[string]*GPUNodeData)
	for _, gpuNode := range ca.gpuNodesByUID {
		gpuNodeByNodeName[gpuNode.NodeName] = gpuNode
	}

	// Create a set of already-removed nodes to avoid duplicates
	removedSet := make(map[string]struct{})
	for _, name := range delta.Remove {
		removedSet[name] = struct{}{}
	}

	// Process dirty nodes
	for nodeName := range ca.dirtyNodes {
		// Skip if already in removal list
		if _, removed := removedSet[nodeName]; removed {
			continue
		}

		// Find node by name
		var node *NodeData
		for _, n := range ca.nodesByUID {
			if n.Name == nodeName {
				node = n
				break
			}
		}

		if node != nil {
			row := ca.buildNodeRow(node, gpuNodeByNodeName)
			delta.Upsert = append(delta.Upsert, row)
		} else {
			delta.Remove = append(delta.Remove, nodeName)
		}
	}

	// Regenerate summary
	snapshot := ca.generateNodeView()
	delta.Summary = &snapshot.Summary

	return delta
}

// generatePendingViewDelta generates delta for pending view
func (ca *ClusterAggregation) generatePendingViewDelta() PendingDelta {
	delta := PendingDelta{
		ByScheduler: make(map[string]int),
	}

	// Include deleted pods in removal list
	delta.Remove = append(delta.Remove, ca.deletedPods...)

	// Process dirty pods
	for uid := range ca.dirtyPods {
		pod := ca.podsByUID[uid]
		if pod == nil {
			// Pod was deleted - already tracked in deletedPods
			continue
		}

		if pod.Phase == "Pending" && pod.NodeName == "" {
			row := ca.buildPendingPodRow(pod)
			delta.Upsert = append(delta.Upsert, row)
		} else {
			// Pod is no longer pending, remove it
			delta.Remove = append(delta.Remove, PodRef{
				Namespace: pod.Namespace,
				Name:      pod.Name,
			})
		}
	}

	// Recalculate scheduler counts
	for _, pod := range ca.podsByUID {
		if pod.Phase == "Pending" && pod.NodeName == "" {
			delta.ByScheduler[pod.SchedulerName]++
		}
	}

	// Recalculate reasons - use message (consistent with snapshot)
	reasonCounts := make(map[string]int)
	for _, pod := range ca.podsByUID {
		if pod.Phase != "Pending" || pod.NodeName != "" {
			continue
		}
		podKey := pod.Namespace + "/" + pod.Name
		if events := ca.eventsByPod[podKey]; len(events) > 0 {
			latest := events[len(events)-1]
			if latest.Reason == ReasonFailedScheduling || latest.Reason == ReasonGPUQuotaOrCapacityNotEnough {
				message := latest.Message
				if len(message) > 200 {
					message = message[:200] + "..."
				}
				reasonCounts[message]++
			}
		}
	}
	for reason, count := range reasonCounts {
		delta.Reasons = append(delta.Reasons, ReasonBucket{Reason: reason, Count: count})
	}

	return delta
}

// generateDRAViewDelta generates delta for DRA view
func (ca *ClusterAggregation) generateDRAViewDelta() DRADelta {
	delta := DRADelta{
		Stats: make(map[string]int),
	}

	// Include deleted claims in removal list
	delta.RemoveClaims = append(delta.RemoveClaims, ca.deletedClaims...)

	// Include deleted slices in removal list
	delta.RemoveSlices = append(delta.RemoveSlices, ca.deletedSlices...)

	// Process dirty claims
	for uid := range ca.dirtyClaims {
		claim := ca.claimsByUID[uid]
		if claim != nil {
			row := ClaimRow{
				Namespace:     claim.Namespace,
				Name:          claim.Name,
				ResourceClass: claim.ResourceClass,
				Driver:        claim.Driver,
				AllocationRaw: claim.Allocation,
				PodRef:        claim.PodRef,
			}
			delta.UpsertClaims = append(delta.UpsertClaims, row)
		}
		// Note: deleted claims are already tracked in deletedClaims
	}

	// Recalculate stats
	for _, claim := range ca.claimsByUID {
		if claim.Allocation != nil {
			delta.Stats["allocated"]++
		} else {
			delta.Stats["pending"]++
		}
	}
	for _, slice := range ca.slicesByUID {
		delta.Stats["totalDevices"] += slice.Devices
	}

	return delta
}

// generateMigrationViewDelta generates delta for migration view
func (ca *ClusterAggregation) generateMigrationViewDelta() MigrationDelta {
	// For migration view, just regenerate the full view
	// as it's relatively small and conflicts need full recalculation
	snapshot := ca.generateMigrationView()

	return MigrationDelta{
		UsedByStats:          snapshot.UsedByStats,
		Conflicts:            snapshot.Conflicts,
		ProgressiveMigration: &snapshot.ProgressiveMigration,
	}
}

// Helper functions

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "< 1m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		if mins > 0 {
			return fmt.Sprintf("%dh%dm", hours, mins)
		}
		return fmt.Sprintf("%dh", hours)
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}
