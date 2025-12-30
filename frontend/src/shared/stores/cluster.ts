import { defineStore } from 'pinia'
import type { ClusterInfo } from '@/lib/k8s-sdk'
import { k8s } from '@/lib/k8s-sdk'

export const useClusterStore = defineStore('cluster', () => {
  // State
  const clusters = ref<Map<string, ClusterInfo>>(new Map())
  const activeClusterId = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const activeCluster = computed(() => {
    if (!activeClusterId.value) return null
    return clusters.value.get(activeClusterId.value) ?? null
  })

  const clusterList = computed(() => Array.from(clusters.value.values()))

  const hasActivecluster = computed(() => activeClusterId.value !== null)

  // Actions
  async function loadClusters() {
    loading.value = true
    error.value = null
    try {
      const result = await k8s.getClusters()
      clusters.value = new Map(Object.entries(result))

      // Set first cluster as active if none selected
      if (!activeClusterId.value && clusters.value.size > 0) {
        activeClusterId.value = clusters.value.keys().next().value ?? null
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      throw e
    } finally {
      loading.value = false
    }
  }

  async function addCluster(name: string, kubeconfig: string, context: string) {
    loading.value = true
    error.value = null
    try {
      const clusterId = await k8s.addCluster(name, kubeconfig, context)
      await loadClusters()
      activeClusterId.value = clusterId
      return clusterId
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      throw e
    } finally {
      loading.value = false
    }
  }

  async function removeCluster(clusterId: string) {
    loading.value = true
    error.value = null
    try {
      await k8s.removeCluster(clusterId)
      clusters.value.delete(clusterId)

      // Switch to another cluster if the active one was removed
      if (activeClusterId.value === clusterId) {
        activeClusterId.value = clusters.value.keys().next().value ?? null
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      throw e
    } finally {
      loading.value = false
    }
  }

  function setActiveCluster(clusterId: string | null) {
    if (clusterId && !clusters.value.has(clusterId)) {
      console.warn(`Cluster ${clusterId} not found`)
      return
    }
    activeClusterId.value = clusterId
  }

  // Event handlers setup
  function setupEventListeners() {
    k8s.onClusterAdded((cluster) => {
      clusters.value.set(cluster.id, cluster)
    })

    k8s.onClusterRemoved((clusterId) => {
      clusters.value.delete(clusterId)
      if (activeClusterId.value === clusterId) {
        activeClusterId.value = clusters.value.keys().next().value ?? null
      }
    })

    k8s.onClusterUpdated((cluster) => {
      clusters.value.set(cluster.id, cluster)
    })
  }

  return {
    // State
    clusters,
    activeClusterId,
    loading,
    error,

    // Getters
    activeCluster,
    clusterList,
    hasActivecluster,

    // Actions
    loadClusters,
    addCluster,
    removeCluster,
    setActiveCluster,
    setupEventListeners
  }
})
