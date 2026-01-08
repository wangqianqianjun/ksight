import type { SchedulerAggregationRequest, SchedulerSnapshot } from '@/plugins/scheduling/types'
import { EventsOn } from '@/wailsjs/runtime/runtime'

// API base URL - will be proxied by Vite in development
const API_BASE = '/api'

async function readJsonBody(response: Response): Promise<{ json: any | null; text: string }> {
  const text = await response.text()
  if (!text) return { json: null, text: '' }
  try {
    return { json: JSON.parse(text), text }
  } catch {
    return { json: null, text }
  }
}

function errorFromResponse(response: Response, json: any | null, text: string): Error {
  const message =
    (json && typeof json === 'object' && 'error' in json && typeof json.error === 'string' && json.error) ||
    text ||
    `${response.status} ${response.statusText}`.trim() ||
    'Request failed'
  return new Error(message)
}

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

export interface NodeSummary {
  name: string
  roles: string
  version: string
  ready: boolean
  age: string
  unschedulable: boolean
}

export interface PodSummary {
  name: string
  namespace: string
  nodeName: string
  phase: string
  age: string
}

type BackendEvent =
  | 'cluster:added'
  | 'cluster:removed'
  | 'cluster:updated'
  | 'resource:event'
  | 'watcher:added'
  | 'watcher:removed'
  | 'scheduler:snapshot'
  | 'scheduler:delta'
  | 'scheduler:warning'

interface BackendAdapter {
  addCluster(name: string, kubeconfig: string, context: string): Promise<string>
  removeCluster(clusterId: string): Promise<void>
  getClusters(): Promise<Record<string, ClusterInfo>>
  toggleClusterPin(clusterId: string): Promise<void>
  addResourceWatcher(request: ResourceWatchRequest): Promise<void>
  removeResourceWatcher(request: Omit<ResourceWatchRequest, 'namespace'>): Promise<void>
  getResourceTypes(clusterId: string): Promise<GroupVersionResource[]>
  loadKubeconfigFromFile(filePath: string): Promise<string>
  saveKubeconfigToFile(content: string, fileName: string): Promise<string>
  getKubeconfigFiles(): Promise<string[]>
  watchDefaultKubeconfig(): Promise<void>
  startSchedulerAggregation(request: SchedulerAggregationRequest): Promise<void>
  stopSchedulerAggregation(clusterId: string): Promise<void>
  getSchedulerSnapshot(clusterId: string): Promise<SchedulerSnapshot>
  getNodes(clusterId: string): Promise<NodeSummary[]>
  getPods(clusterId: string, namespace?: string): Promise<PodSummary[]>
  loadDefaultKubeconfig(): Promise<string>
  on(event: BackendEvent, callback: (data: any) => void): () => void
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

const isWailsAvailable = () => {
  return typeof window !== 'undefined' && !!window.go?.main?.App
}

const getWailsApp = () => {
  const app = window.go?.main?.App
  if (!app) {
    throw new Error('Wails bindings not available')
  }
  return app
}

const createWailsAdapter = (): BackendAdapter => ({
  addCluster: (name, kubeconfig, context) => getWailsApp().AddCluster(name, kubeconfig, context),
  removeCluster: (clusterId) => getWailsApp().RemoveCluster(clusterId),
  getClusters: () => getWailsApp().GetClusters(),
  toggleClusterPin: (clusterId) => getWailsApp().ToggleClusterPin(clusterId),
  addResourceWatcher: (request) =>
    getWailsApp().AddResourceWatcher(
      request.clusterId,
      request.group,
      request.version,
      request.resource,
      request.namespace ?? ''
    ),
  removeResourceWatcher: (request) =>
    getWailsApp().RemoveResourceWatcher(
      request.clusterId,
      request.group,
      request.version,
      request.resource
    ),
  getResourceTypes: (clusterId) => getWailsApp().GetResourceTypes(clusterId),
  loadKubeconfigFromFile: (filePath) => getWailsApp().LoadKubeconfigFromFile(filePath),
  saveKubeconfigToFile: (content, fileName) => getWailsApp().SaveKubeconfigToFile(content, fileName),
  getKubeconfigFiles: () => getWailsApp().GetKubeconfigFiles(),
  watchDefaultKubeconfig: () => getWailsApp().WatchDefaultKubeconfig(),
  startSchedulerAggregation: (request) => getWailsApp().StartSchedulerAggregation(request),
  stopSchedulerAggregation: (clusterId) => getWailsApp().StopSchedulerAggregation(clusterId),
  getSchedulerSnapshot: (clusterId) => getWailsApp().GetSchedulerSnapshot(clusterId),
  getNodes: (clusterId) => getWailsApp().GetNodes(clusterId),
  getPods: (clusterId, namespace) => getWailsApp().GetPods(clusterId, namespace ?? ''),
  loadDefaultKubeconfig: () => getWailsApp().LoadDefaultKubeconfig(),
  on: (event, callback) => EventsOn(event, (payload: any) => callback(payload)),
})

const createHttpAdapter = (): BackendAdapter => {
  const wsManager = new WebSocketManager()

  if (typeof window !== 'undefined') {
    wsManager.connect()
  }

  return {
    async addCluster(name: string, kubeconfig: string, context: string = ''): Promise<string> {
      const response = await fetch(`${API_BASE}/clusters`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, kubeconfig, context })
      })
      const { json, text } = await readJsonBody(response)
      if (!response.ok) throw errorFromResponse(response, json, text)
      if (json && typeof json === 'object' && typeof json.clusterId === 'string') return json.clusterId
      throw new Error('Invalid response from server')
    },

    async removeCluster(clusterId: string): Promise<void> {
      const response = await fetch(`${API_BASE}/clusters/${clusterId}`, {
        method: 'DELETE'
      })
      if (!response.ok) {
        const { json, text } = await readJsonBody(response)
        throw errorFromResponse(response, json, text)
      }
    },

    async getClusters(): Promise<Record<string, ClusterInfo>> {
      const response = await fetch(`${API_BASE}/clusters`)
      const { json, text } = await readJsonBody(response)
      if (!response.ok) throw errorFromResponse(response, json, text)
      if (json && typeof json === 'object') return json as Record<string, ClusterInfo>
      throw new Error('Invalid response from server')
    },

    async toggleClusterPin(clusterId: string): Promise<void> {
      // Not implemented in HTTP API yet
      console.warn('toggleClusterPin not implemented')
    },

    async addResourceWatcher(request: ResourceWatchRequest): Promise<void> {
      // Not implemented in HTTP API yet - resources are watched via scheduler aggregation
      console.warn('addResourceWatcher not implemented')
    },

    async removeResourceWatcher(request: Omit<ResourceWatchRequest, 'namespace'>): Promise<void> {
      // Not implemented in HTTP API yet
      console.warn('removeResourceWatcher not implemented')
    },

    async getResourceTypes(clusterId: string): Promise<GroupVersionResource[]> {
      // Not implemented in HTTP API yet
      console.warn('getResourceTypes not implemented')
      return []
    },

    async loadKubeconfigFromFile(filePath: string): Promise<string> {
      // Not implemented in HTTP API yet
      console.warn('loadKubeconfigFromFile not implemented')
      return ''
    },

    async saveKubeconfigToFile(content: string, fileName: string): Promise<string> {
      // Not implemented in HTTP API yet
      console.warn('saveKubeconfigToFile not implemented')
      return ''
    },

    async getKubeconfigFiles(): Promise<string[]> {
      // Not implemented in HTTP API yet
      console.warn('getKubeconfigFiles not implemented')
      return []
    },

    async watchDefaultKubeconfig(): Promise<void> {
      // Not implemented in HTTP API yet
      console.warn('watchDefaultKubeconfig not implemented')
    },

    async startSchedulerAggregation(request: SchedulerAggregationRequest): Promise<void> {
      const response = await fetch(`${API_BASE}/scheduler/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request)
      })
      if (!response.ok) {
        const { json, text } = await readJsonBody(response)
        throw errorFromResponse(response, json, text)
      }
    },

    async stopSchedulerAggregation(clusterId: string): Promise<void> {
      const response = await fetch(`${API_BASE}/scheduler/stop`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ clusterId })
      })
      if (!response.ok) {
        const { json, text } = await readJsonBody(response)
        throw errorFromResponse(response, json, text)
      }
    },

    async getSchedulerSnapshot(clusterId: string): Promise<SchedulerSnapshot> {
      const response = await fetch(`${API_BASE}/scheduler/snapshot/${clusterId}`)
      const { json, text } = await readJsonBody(response)
      if (!response.ok) throw errorFromResponse(response, json, text)
      if (json) return json as SchedulerSnapshot
      throw new Error('Invalid response from server')
    },

    async getNodes(clusterId: string): Promise<NodeSummary[]> {
      const response = await fetch(`${API_BASE}/clusters/${clusterId}/nodes`)
      if (!response.ok) {
        const errText = await response.text()
        throw new Error(errText || 'Failed to get nodes')
      }
      return response.json()
    },

    async getPods(clusterId: string, namespace?: string): Promise<PodSummary[]> {
      const query = namespace ? `?namespace=${encodeURIComponent(namespace)}` : ''
      const response = await fetch(`${API_BASE}/clusters/${clusterId}/pods${query}`)
      if (!response.ok) {
        const errText = await response.text()
        throw new Error(errText || 'Failed to get pods')
      }
      return response.json()
    },

    async loadDefaultKubeconfig(): Promise<string> {
      const response = await fetch(`${API_BASE}/clusters/load-default`, { method: 'POST' })
      const { json, text } = await readJsonBody(response)
      if (!response.ok) throw errorFromResponse(response, json, text)
      if (json && typeof json === 'object' && typeof json.clusterId === 'string') return json.clusterId
      throw new Error('Invalid response from server')
    },

    on: (event, callback) => wsManager.on(event, callback),
  }
}

export class K8sSDK {
  private adapter: BackendAdapter

  constructor(adapter?: BackendAdapter) {
    this.adapter = adapter ?? (isWailsAvailable() ? createWailsAdapter() : createHttpAdapter())
  }

  addCluster(name: string, kubeconfig: string, context: string = ''): Promise<string> {
    return this.adapter.addCluster(name, kubeconfig, context)
  }

  removeCluster(clusterId: string): Promise<void> {
    return this.adapter.removeCluster(clusterId)
  }

  getClusters(): Promise<Record<string, ClusterInfo>> {
    return this.adapter.getClusters()
  }

  toggleClusterPin(clusterId: string): Promise<void> {
    return this.adapter.toggleClusterPin(clusterId)
  }

  addResourceWatcher(request: ResourceWatchRequest): Promise<void> {
    return this.adapter.addResourceWatcher(request)
  }

  removeResourceWatcher(request: Omit<ResourceWatchRequest, 'namespace'>): Promise<void> {
    return this.adapter.removeResourceWatcher(request)
  }

  getResourceTypes(clusterId: string): Promise<GroupVersionResource[]> {
    return this.adapter.getResourceTypes(clusterId)
  }

  loadKubeconfigFromFile(filePath: string): Promise<string> {
    return this.adapter.loadKubeconfigFromFile(filePath)
  }

  saveKubeconfigToFile(content: string, fileName: string): Promise<string> {
    return this.adapter.saveKubeconfigToFile(content, fileName)
  }

  getKubeconfigFiles(): Promise<string[]> {
    return this.adapter.getKubeconfigFiles()
  }

  watchDefaultKubeconfig(): Promise<void> {
    return this.adapter.watchDefaultKubeconfig()
  }

  startSchedulerAggregation(request: SchedulerAggregationRequest): Promise<void> {
    return this.adapter.startSchedulerAggregation(request)
  }

  stopSchedulerAggregation(clusterId: string): Promise<void> {
    return this.adapter.stopSchedulerAggregation(clusterId)
  }

  getSchedulerSnapshot(clusterId: string): Promise<SchedulerSnapshot> {
    return this.adapter.getSchedulerSnapshot(clusterId)
  }

  getNodes(clusterId: string): Promise<NodeSummary[]> {
    return this.adapter.getNodes(clusterId)
  }

  getPods(clusterId: string, namespace?: string): Promise<PodSummary[]> {
    return this.adapter.getPods(clusterId, namespace)
  }

  loadDefaultKubeconfig(): Promise<string> {
    return this.adapter.loadDefaultKubeconfig()
  }

  onClusterAdded(callback: (cluster: ClusterInfo) => void): () => void {
    return this.adapter.on('cluster:added', callback)
  }

  onClusterRemoved(callback: (clusterId: string) => void): () => void {
    return this.adapter.on('cluster:removed', callback)
  }

  onClusterUpdated(callback: (cluster: ClusterInfo) => void): () => void {
    return this.adapter.on('cluster:updated', callback)
  }

  onResourceEvent(callback: (event: ResourceEvent) => void): () => void {
    return this.adapter.on('resource:event', callback)
  }

  // Scheduler event listeners (HTTP path only currently)
  onSchedulerSnapshot(callback: (snapshot: SchedulerSnapshot) => void): () => void {
    return this.adapter.on('scheduler:snapshot', callback)
  }

  onSchedulerDelta(callback: (delta: any) => void): () => void {
    return this.adapter.on('scheduler:delta', callback)
  }

  onSchedulerWarning(callback: (warning: any) => void): () => void {
    return this.adapter.on('scheduler:warning', callback)
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
