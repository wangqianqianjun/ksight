import { ref } from 'vue'

// Global state for the cluster connection dialog
const isOpen = ref(false)

export function useClusterDialog() {
  function open() {
    isOpen.value = true
  }

  function close() {
    isOpen.value = false
  }

  function toggle() {
    isOpen.value = !isOpen.value
  }

  return {
    isOpen,
    open,
    close,
    toggle
  }
}
