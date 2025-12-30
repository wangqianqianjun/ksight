import type { SchedulerAggregationRequest, SchedulerSnapshot } from '@/plugins/scheduling/types'

// API base URL - will be proxied by Vite in development
const API_BASE = '/api'

// Types for cluster management
export interface ClusterInfo {
  id: string
  name: string
  context: string
  server: string
  status: 'connected' | 'disconnected' | 'error'
  lastError?: string
  isPinned: boolean
}

export interface ResourceWatchRequest {
  clusterId: string
  group: string
  version: string
  resource: string
  namespace?: string
}

export interface GroupVersionResource {
  group: string
  version: string
  resource: string
}

export interface ResourceEvent {
  type: 'ADDED' | 'MODIFIED' | 'DELETED'
  clusterId: string
  gvr: GroupVersionResource
  namespace: string
  name: string
  object: any
  oldObject?: any
  timestamp: string
}

// WebSocket event message
interface WSMessage {
  event: string
  data: any
}

// WebSocket Manager for real-time events
class WebSocketManager {
  private ws: WebSocket | null = null
  private eventListeners: Map<string, Set<Function>> = new Map()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/ws`

    try {
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        console.log('WebSocket connected')
        this.reconnectAttempts = 0
      }

      this.ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data)
          this.dispatchEvent(msg.event, msg.data)
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e)
        }
      }

      this.ws.onclose = () => {
        console.log('WebSocket disconnected')
        this.scheduleReconnect()
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
      }
    } catch (e) {
      console.error('Failed to create WebSocket:', e)
      this.scheduleReconnect()
    }
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      const delay = this.reconnectDelay * this.reconnectAttempts
      console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`)
      setTimeout(() => this.connect(), delay)
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  on(event: string, callback: Function): () => void {
    if (!this.eventListeners.has(event)) {
      this.eventListeners.set(event, new Set())
    }
    this.eventListeners.get(event)!.add(callback)

    return () => {
      const listeners = this.eventListeners.get(event)
      if (listeners) {
        listeners.delete(callback)
      }
    }
  }

  private dispatchEvent(event: string, data: any) {
    const listeners = this.eventListeners.get(event)
    if (listeners) {
      listeners.forEach(callback => {
        try {
          callback(data)
        } catch (e) {
          console.error(`Error in event listener for ${event}:`, e)
        }
      })
    }
  }
}

// Global WebSocket manager
export const wsManager = new WebSocketManager()

// K8s SDK Class using HTTP API
export class K8sSDK {
  private eventListeners: Map<string, Set<Function>> = new Map()

  constructor() {
    // Connect WebSocket when SDK is created
    if (typeof window !== 'undefined') {
      wsManager.connect()
    }
  }

  async addCluster(name: string, kubeconfig: string, context: string = ''): Promise<string> {
    const response = await fetch(`${API_BASE}/clusters`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, kubeconfig, context })
    })
    const data = await response.json()
    if (!response.ok) throw new Error(data.error || 'Failed to add cluster')
    return data.clusterId
  }

  async removeCluster(clusterId: string): Promise<void> {
    const response = await fetch(`${API_BASE}/clusters/${clusterId}`, {
      method: 'DELETE'
    })
    if (!response.ok) {
      const data = await response.json()
      throw new Error(data.error || 'Failed to remove cluster')
    }
  }

  async getClusters(): Promise<Record<string, ClusterInfo>> {
    const response = await fetch(`${API_BASE}/clusters`)
    const data = await response.json()
    if (!response.ok) throw new Error(data.error || 'Failed to get clusters')
    return data
  }

  async toggleClusterPin(clusterId: string): Promise<void> {
    // Not implemented in HTTP API yet
    console.warn('toggleClusterPin not implemented')
  }

  async addResourceWatcher(request: ResourceWatchRequest): Promise<void> {
    // Not implemented in HTTP API yet - resources are watched via scheduler aggregation
    console.warn('addResourceWatcher not implemented')
  }

  async removeResourceWatcher(request: Omit<ResourceWatchRequest, 'namespace'>): Promise<void> {
    // Not implemented in HTTP API yet
    console.warn('removeResourceWatcher not implemented')
  }

  async getResourceTypes(clusterId: string): Promise<GroupVersionResource[]> {
    // Not implemented in HTTP API yet
    console.warn('getResourceTypes not implemented')
    return []
  }

  async loadKubeconfigFromFile(filePath: string): Promise<string> {
    // Not implemented in HTTP API yet
    console.warn('loadKubeconfigFromFile not implemented')
    return ''
  }

  async saveKubeconfigToFile(content: string, fileName: string): Promise<string> {
    // Not implemented in HTTP API yet
    console.warn('saveKubeconfigToFile not implemented')
    return ''
  }

  async getKubeconfigFiles(): Promise<string[]> {
    // Not implemented in HTTP API yet
    console.warn('getKubeconfigFiles not implemented')
    return []
  }

  async watchDefaultKubeconfig(): Promise<void> {
    // Not implemented in HTTP API yet
    console.warn('watchDefaultKubeconfig not implemented')
  }

  // Scheduler aggregation methods
  async startSchedulerAggregation(request: SchedulerAggregationRequest): Promise<void> {
    const response = await fetch(`${API_BASE}/scheduler/start`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request)
    })
    if (!response.ok) {
      const data = await response.json()
      throw new Error(data.error || 'Failed to start scheduler aggregation')
    }
  }

  async stopSchedulerAggregation(clusterId: string): Promise<void> {
    const response = await fetch(`${API_BASE}/scheduler/stop`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ clusterId })
    })
    if (!response.ok) {
      const data = await response.json()
      throw new Error(data.error || 'Failed to stop scheduler aggregation')
    }
  }

  async getSchedulerSnapshot(clusterId: string): Promise<SchedulerSnapshot> {
    const response = await fetch(`${API_BASE}/scheduler/snapshot/${clusterId}`)
    const data = await response.json()
    if (!response.ok) throw new Error(data.error || 'Failed to get scheduler snapshot')
    return data
  }

  // Event listeners using WebSocket
  onClusterAdded(callback: (cluster: ClusterInfo) => void): () => void {
    return wsManager.on('cluster:added', callback)
  }

  onClusterRemoved(callback: (clusterId: string) => void): () => void {
    return wsManager.on('cluster:removed', callback)
  }

  onClusterUpdated(callback: (cluster: ClusterInfo) => void): () => void {
    return wsManager.on('cluster:updated', callback)
  }

  onResourceEvent(callback: (event: ResourceEvent) => void): () => void {
    return wsManager.on('resource:event', callback)
  }

  // Scheduler event listeners
  onSchedulerSnapshot(callback: (snapshot: SchedulerSnapshot) => void): () => void {
    return wsManager.on('scheduler:snapshot', callback)
  }

  onSchedulerDelta(callback: (delta: any) => void): () => void {
    return wsManager.on('scheduler:delta', callback)
  }

  onSchedulerWarning(callback: (warning: any) => void): () => void {
    return wsManager.on('scheduler:warning', callback)
  }

  get pods() {
    return { group: '', version: 'v1', resource: 'pods' }
  }

  get deployments() {
    return { group: 'apps', version: 'v1', resource: 'deployments' }
  }

  get services() {
    return { group: '', version: 'v1', resource: 'services' }
  }

  get nodes() {
    return { group: '', version: 'v1', resource: 'nodes' }
  }
}

// Global SDK instance
export const k8s = new K8sSDK()

// Make it available globally as 'k' for operations
declare global {
  interface Window {
    k: K8sSDK
  }
}

if (typeof window !== 'undefined') {
  window.k = k8s
}
