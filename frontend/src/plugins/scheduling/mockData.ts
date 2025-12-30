import type { SchedulerSnapshot, SchedulerWarning, NodeViewSnapshot, DRASnapshot, MigrationSnapshot, PendingViewSnapshot, SchedulerHealthSnapshot } from './types'

const mockClusterId = 'mock-scheduling'

export function createMockSchedulerSnapshot(now: Date = new Date()): SchedulerSnapshot {
  const generatedAt = now.toISOString()

  const nodeView: NodeViewSnapshot = {
    nodes: [
      {
        name: 'worker-gpu-1',
        zone: 'us-east-1a',
        labels: {
          'topology.kubernetes.io/zone': 'us-east-1a',
          'node.kubernetes.io/instance-type': 'p4d.24xlarge'
        },
        capacity: { 'nvidia.com/gpu': '8' },
        allocatable: { 'nvidia.com/gpu': '7' },
        tensorFusion: {
          pool: 'tf-pool-a',
          phase: 'Running',
          managedGpus: 4
        },
        gpu: {
          total: 8,
          usedBy: {
            'tensor-fusion': 5,
            'nvidia-device-plugin': 3
          },
          devices: [
            {
              name: 'gpu-0',
              model: 'H100',
              usedBy: 'tensor-fusion',
              capacity: { tflops: '120', vram: '80Gi' },
              available: { vram: '20Gi' },
              runningApps: [
                { namespace: 'ml', name: 'tf-train-1' },
                { namespace: 'ml', name: 'tf-train-1', count: 2 }
              ]
            },
            {
              name: 'gpu-1',
              model: 'H100',
              usedBy: 'tensor-fusion',
              capacity: { tflops: '120', vram: '80Gi' },
              available: { vram: '30Gi' },
              runningApps: [{ namespace: 'ml', name: 'tf-train-2' }]
            },
            {
              name: 'gpu-2',
              model: 'H100',
              usedBy: 'tensor-fusion',
              capacity: { tflops: '120', vram: '80Gi' },
              available: { vram: '50Gi' }
            },
            {
              name: 'gpu-3',
              model: 'H100',
              usedBy: 'tensor-fusion',
              capacity: { tflops: '120', vram: '80Gi' },
              available: { vram: '70Gi' }
            },
            {
              name: 'gpu-4',
              model: 'H100',
              usedBy: 'nvidia-device-plugin',
              capacity: { tflops: '120', vram: '80Gi' },
              available: { vram: '10Gi' },
              runningApps: [{ namespace: 'ml', name: 'cuda-train-1' }]
            },
            {
              name: 'gpu-5',
              model: 'H100',
              usedBy: 'nvidia-device-plugin',
              capacity: { tflops: '120', vram: '80Gi' },
              available: { vram: '15Gi' }
            },
            {
              name: 'gpu-6',
              model: 'H100',
              usedBy: 'nvidia-device-plugin',
              capacity: { tflops: '120', vram: '80Gi' },
              available: { vram: '22Gi' }
            },
            {
              name: 'gpu-7',
              model: 'H100',
              usedBy: 'nvidia-device-plugin',
              capacity: { tflops: '120', vram: '80Gi' },
              available: { vram: '40Gi' }
            }
          ]
        },
        pods: [
          { namespace: 'ml', name: 'tf-train-1' },
          { namespace: 'ml', name: 'cuda-train-1' }
        ]
      },
      {
        name: 'worker-gpu-2',
        zone: 'us-east-1b',
        labels: {
          'topology.kubernetes.io/zone': 'us-east-1b',
          'node.kubernetes.io/instance-type': 'g5.12xlarge'
        },
        capacity: { 'nvidia.com/gpu': '4' },
        allocatable: { 'nvidia.com/gpu': '4' },
        tensorFusion: {
          pool: 'tf-pool-b',
          phase: 'Pending',
          managedGpus: 0
        },
        gpu: {
          total: 4,
          usedBy: {
            'nvidia-device-plugin': 4
          },
          devices: [
            {
              name: 'gpu-0',
              model: 'A10G',
              usedBy: 'nvidia-device-plugin',
              capacity: { tflops: '60', vram: '24Gi' },
              available: { vram: '2Gi' },
              runningApps: [{ namespace: 'ml', name: 'cuda-train-2' }]
            },
            {
              name: 'gpu-1',
              model: 'A10G',
              usedBy: 'nvidia-device-plugin',
              capacity: { tflops: '60', vram: '24Gi' },
              available: { vram: '5Gi' }
            },
            {
              name: 'gpu-2',
              model: 'A10G',
              usedBy: 'nvidia-device-plugin',
              capacity: { tflops: '60', vram: '24Gi' },
              available: { vram: '6Gi' }
            },
            {
              name: 'gpu-3',
              model: 'A10G',
              usedBy: 'nvidia-device-plugin',
              capacity: { tflops: '60', vram: '24Gi' },
              available: { vram: '4Gi' }
            }
          ]
        }
      },
      {
        name: 'worker-gpu-3',
        zone: 'us-east-1b',
        labels: {
          'topology.kubernetes.io/zone': 'us-east-1b',
          'node.kubernetes.io/instance-type': 'g4dn.12xlarge'
        },
        capacity: { 'nvidia.com/gpu': '2' },
        allocatable: { 'nvidia.com/gpu': '2' },
        tensorFusion: {
          pool: 'tf-pool-a',
          phase: 'Migrating',
          managedGpus: 2
        },
        gpu: {
          total: 2,
          usedBy: {
            unknown: 2
          },
          devices: [
            {
              name: 'gpu-0',
              model: 'T4',
              usedBy: 'unknown',
              capacity: { tflops: '16', vram: '16Gi' },
              available: { vram: '16Gi' }
            },
            {
              name: 'gpu-1',
              model: 'T4',
              usedBy: 'unknown',
              capacity: { tflops: '16', vram: '16Gi' },
              available: { vram: '16Gi' }
            }
          ]
        }
      }
    ],
    summary: {
      totalNodes: 3,
      readyNodes: 2,
      totalGpus: 14,
      gpusByUsedBy: {
        'tensor-fusion': 5,
        'nvidia-device-plugin': 7,
        unknown: 2
      },
      gpusByModel: {
        H100: 8,
        A10G: 4,
        T4: 2
      },
      totalTflops: 920,
      availableTflops: 180,
      totalVram: 656,
      availableVram: 148
    }
  }

  const pendingView: PendingViewSnapshot = {
    pods: [
      {
        namespace: 'ml',
        name: 'tf-train-7',
        scheduler: 'tensor-fusion-scheduler',
        gpuRequest: '4',
        pool: 'tf-pool-a',
        reason: '0/6 nodes are available: 2 Insufficient nvidia.com/gpu',
        since: '42m'
      },
      {
        namespace: 'ml',
        name: 'tf-train-9',
        scheduler: 'tensor-fusion-scheduler',
        gpuRequest: '8',
        pool: 'tf-pool-a',
        reason: 'nvidia.com/gpu: preferred node has taint tensor-fusion/pool=dedicated',
        since: '2h'
      },
      {
        namespace: 'research',
        name: 'cuda-train-5',
        scheduler: 'default-scheduler',
        gpuRequest: '2',
        pool: 'native',
        reason: 'Insufficient memory',
        since: '18m'
      },
      {
        namespace: 'batch',
        name: 'cuda-batch-1',
        scheduler: 'default-scheduler',
        gpuRequest: '1',
        pool: 'native',
        since: '6m'
      }
    ],
    byScheduler: {
      'tensor-fusion-scheduler': 2,
      'default-scheduler': 2
    },
    reasons: [
      { reason: 'Insufficient nvidia.com/gpu', count: 2 },
      { reason: 'Insufficient memory', count: 1 },
      { reason: 'Taints not tolerated', count: 1 }
    ],
    eventsAvailable: true
  }

  const draView: DRASnapshot = {
    claims: [
      {
        namespace: 'ml',
        name: 'claim-a',
        resourceClass: 'gpu.tensor-fusion.ai',
        driver: 'tensor-fusion.ai/driver',
        allocation: {
          devices: ['gpu-0', 'gpu-1'],
          slice: 'slice-a'
        },
        podRef: { namespace: 'ml', name: 'tf-train-1' }
      },
      {
        namespace: 'ml',
        name: 'claim-b',
        resourceClass: 'gpu.tensor-fusion.ai',
        driver: 'tensor-fusion.ai/driver',
        allocation: {
          devices: ['gpu-4'],
          slice: 'slice-b'
        },
        podRef: { namespace: 'ml', name: 'tf-train-2' }
      },
      {
        namespace: 'research',
        name: 'claim-c',
        resourceClass: 'gpu.tensor-fusion.ai',
        driver: 'tensor-fusion.ai/driver'
      }
    ],
    slices: [
      {
        name: 'slice-a',
        nodeName: 'worker-gpu-1',
        driver: 'tensor-fusion.ai/driver',
        devices: 4
      },
      {
        name: 'slice-b',
        nodeName: 'worker-gpu-1',
        driver: 'tensor-fusion.ai/driver',
        devices: 2
      },
      {
        name: 'slice-c',
        nodeName: 'worker-gpu-2',
        driver: 'tensor-fusion.ai/driver',
        devices: 4
      }
    ],
    stats: {
      unallocated: 1,
      allocating: 1,
      allocated: 2,
      totalDevices: 10
    },
    status: 'available' as const
  }

  const migrationView: MigrationSnapshot = {
    usedByStats: {
      'tensor-fusion': 5,
      'nvidia-device-plugin': 7,
      unknown: 2
    },
    conflicts: [
      {
        node: 'worker-gpu-3',
        pod: { namespace: 'research', name: 'cuda-train-5' },
        usedBy: 'tensor-fusion',
        resourceType: 'nvidia.com/gpu'
      },
      {
        node: 'worker-gpu-1',
        pod: { namespace: 'ml', name: 'tf-train-9' },
        usedBy: 'nvidia-device-plugin',
        resourceType: 'tensor-fusion'
      }
    ],
    progressiveMigration: true
  }

  const health: SchedulerHealthSnapshot = {
    totalPending: pendingView.pods.length,
    totalNodes: nodeView.nodes.length,
    totalGpus: nodeView.summary.totalGpus,
    warnings: ['GPU_USED_BY_MISMATCH']
  }

  return {
    clusterId: mockClusterId,
    generatedAt,
    seq: 1,
    nodeView,
    pendingView,
    draView,
    migrationView,
    health
  }
}

export function createMockSchedulerWarnings(clusterId: string = mockClusterId): SchedulerWarning[] {
  return [
    {
      clusterId,
      generatedAt: new Date().toISOString(),
      type: 'GPU_USED_BY_MISMATCH',
      message: 'GPU usedBy does not match pod resource type on worker-gpu-3.',
      objects: [
        { kind: 'Node', name: 'worker-gpu-3' },
        { kind: 'Pod', namespace: 'research', name: 'cuda-train-5' }
      ]
    },
    {
      clusterId,
      generatedAt: new Date().toISOString(),
      type: 'SCHEDULER_MISMATCH',
      message: 'Pod scheduled with tensor-fusion-scheduler requests nvidia.com/gpu.',
      objects: [{ kind: 'Pod', namespace: 'ml', name: 'tf-train-9' }]
    }
  ]
}

export function getMockClusterId(): string {
  return mockClusterId
}
