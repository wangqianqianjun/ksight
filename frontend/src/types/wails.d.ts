import type { SchedulerAggregationRequest, SchedulerSnapshot } from '@/plugins/scheduling/types'
import type {
  ClusterInfo,
  GroupVersionResource,
  NodeSummary,
  PodSummary,
} from '@/lib/k8s-sdk'

export {}

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          AddCluster(name: string, kubeconfig: string, context: string): Promise<string>
          RemoveCluster(clusterId: string): Promise<void>
          GetClusters(): Promise<Record<string, ClusterInfo>>
          ToggleClusterPin(clusterId: string): Promise<void>
          AddResourceWatcher(clusterId: string, group: string, version: string, resource: string, namespace: string): Promise<void>
          RemoveResourceWatcher(clusterId: string, group: string, version: string, resource: string): Promise<void>
          GetResourceTypes(clusterId: string): Promise<GroupVersionResource[]>
          LoadKubeconfigFromFile(filePath: string): Promise<string>
          SaveKubeconfigToFile(content: string, fileName: string): Promise<string>
          GetKubeconfigFiles(): Promise<string[]>
          WatchDefaultKubeconfig(): Promise<void>
          StartSchedulerAggregation(request: SchedulerAggregationRequest): Promise<void>
          StopSchedulerAggregation(clusterId: string): Promise<void>
          GetSchedulerSnapshot(clusterId: string): Promise<SchedulerSnapshot>
          GetNodes(clusterId: string): Promise<NodeSummary[]>
          GetPods(clusterId: string, namespace: string): Promise<PodSummary[]>
          LoadDefaultKubeconfig(): Promise<string>
        }
      }
    }
  }
}
