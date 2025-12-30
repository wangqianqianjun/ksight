export namespace scheduler {
	
	export class AggregationIncludeConfig {
	    pods?: boolean;
	    nodes?: boolean;
	    events?: boolean;
	    dra?: boolean;
	    tensorFusion?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AggregationIncludeConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pods = source["pods"];
	        this.nodes = source["nodes"];
	        this.events = source["events"];
	        this.dra = source["dra"];
	        this.tensorFusion = source["tensorFusion"];
	    }
	}
	export class AppRef {
	    namespace: string;
	    name: string;
	    count?: number;
	
	    static createFrom(source: any = {}) {
	        return new AppRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.namespace = source["namespace"];
	        this.name = source["name"];
	        this.count = source["count"];
	    }
	}
	export class PodRef {
	    namespace: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new PodRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.namespace = source["namespace"];
	        this.name = source["name"];
	    }
	}
	export class ClaimRow {
	    namespace: string;
	    name: string;
	    allocation?: Record<string, any>;
	    resourceClass?: string;
	    driver?: string;
	    podRef?: PodRef;
	
	    static createFrom(source: any = {}) {
	        return new ClaimRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.namespace = source["namespace"];
	        this.name = source["name"];
	        this.allocation = source["allocation"];
	        this.resourceClass = source["resourceClass"];
	        this.driver = source["driver"];
	        this.podRef = this.convertValues(source["podRef"], PodRef);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConflictRow {
	    node: string;
	    pod: PodRef;
	    usedBy: string;
	    resourceType: string;
	
	    static createFrom(source: any = {}) {
	        return new ConflictRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.node = source["node"];
	        this.pod = this.convertValues(source["pod"], PodRef);
	        this.usedBy = source["usedBy"];
	        this.resourceType = source["resourceType"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SliceRow {
	    name: string;
	    nodeName?: string;
	    driver: string;
	    devices: number;
	
	    static createFrom(source: any = {}) {
	        return new SliceRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.nodeName = source["nodeName"];
	        this.driver = source["driver"];
	        this.devices = source["devices"];
	    }
	}
	export class DRASnapshot {
	    claims: ClaimRow[];
	    slices: SliceRow[];
	    stats?: Record<string, number>;
	    status?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new DRASnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.claims = this.convertValues(source["claims"], ClaimRow);
	        this.slices = this.convertValues(source["slices"], SliceRow);
	        this.stats = source["stats"];
	        this.status = source["status"];
	        this.message = source["message"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GPUDeviceRow {
	    name: string;
	    model?: string;
	    usedBy?: string;
	    capacity?: Record<string, string>;
	    available?: Record<string, string>;
	    runningApps?: AppRef[];
	
	    static createFrom(source: any = {}) {
	        return new GPUDeviceRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.model = source["model"];
	        this.usedBy = source["usedBy"];
	        this.capacity = source["capacity"];
	        this.available = source["available"];
	        this.runningApps = this.convertValues(source["runningApps"], AppRef);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MigrationSnapshot {
	    usedByStats: Record<string, number>;
	    conflicts?: ConflictRow[];
	    progressiveMigration: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MigrationSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.usedByStats = source["usedByStats"];
	        this.conflicts = this.convertValues(source["conflicts"], ConflictRow);
	        this.progressiveMigration = source["progressiveMigration"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class NodeGPUInfo {
	    total: number;
	    usedBy: Record<string, number>;
	    devices: GPUDeviceRow[];
	
	    static createFrom(source: any = {}) {
	        return new NodeGPUInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.usedBy = source["usedBy"];
	        this.devices = this.convertValues(source["devices"], GPUDeviceRow);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class NodeTensorFusionInfo {
	    pool?: string;
	    phase?: string;
	    managedGpus?: number;
	
	    static createFrom(source: any = {}) {
	        return new NodeTensorFusionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pool = source["pool"];
	        this.phase = source["phase"];
	        this.managedGpus = source["managedGpus"];
	    }
	}
	export class NodeRow {
	    name: string;
	    zone?: string;
	    labels?: Record<string, string>;
	    capacity?: Record<string, string>;
	    allocatable?: Record<string, string>;
	    tensorFusion?: NodeTensorFusionInfo;
	    gpu: NodeGPUInfo;
	    pods?: PodRef[];
	
	    static createFrom(source: any = {}) {
	        return new NodeRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.zone = source["zone"];
	        this.labels = source["labels"];
	        this.capacity = source["capacity"];
	        this.allocatable = source["allocatable"];
	        this.tensorFusion = this.convertValues(source["tensorFusion"], NodeTensorFusionInfo);
	        this.gpu = this.convertValues(source["gpu"], NodeGPUInfo);
	        this.pods = this.convertValues(source["pods"], PodRef);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class NodeSummary {
	    totalNodes: number;
	    readyNodes: number;
	    totalGpus: number;
	    gpusByUsedBy?: Record<string, number>;
	    gpusByModel?: Record<string, number>;
	    totalTflops?: number;
	    availableTflops?: number;
	    totalVram?: number;
	    availableVram?: number;
	
	    static createFrom(source: any = {}) {
	        return new NodeSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalNodes = source["totalNodes"];
	        this.readyNodes = source["readyNodes"];
	        this.totalGpus = source["totalGpus"];
	        this.gpusByUsedBy = source["gpusByUsedBy"];
	        this.gpusByModel = source["gpusByModel"];
	        this.totalTflops = source["totalTflops"];
	        this.availableTflops = source["availableTflops"];
	        this.totalVram = source["totalVram"];
	        this.availableVram = source["availableVram"];
	    }
	}
	
	export class NodeViewSnapshot {
	    nodes: NodeRow[];
	    summary: NodeSummary;
	
	    static createFrom(source: any = {}) {
	        return new NodeViewSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodes = this.convertValues(source["nodes"], NodeRow);
	        this.summary = this.convertValues(source["summary"], NodeSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PendingPodRow {
	    namespace: string;
	    name: string;
	    scheduler?: string;
	    gpuRequest?: string;
	    pool?: string;
	    reason?: string;
	    since?: string;
	
	    static createFrom(source: any = {}) {
	        return new PendingPodRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.namespace = source["namespace"];
	        this.name = source["name"];
	        this.scheduler = source["scheduler"];
	        this.gpuRequest = source["gpuRequest"];
	        this.pool = source["pool"];
	        this.reason = source["reason"];
	        this.since = source["since"];
	    }
	}
	export class ReasonBucket {
	    reason: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new ReasonBucket(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reason = source["reason"];
	        this.count = source["count"];
	    }
	}
	export class PendingViewSnapshot {
	    pods: PendingPodRow[];
	    byScheduler: Record<string, number>;
	    reasons?: ReasonBucket[];
	    eventsAvailable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PendingViewSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pods = this.convertValues(source["pods"], PendingPodRow);
	        this.byScheduler = source["byScheduler"];
	        this.reasons = this.convertValues(source["reasons"], ReasonBucket);
	        this.eventsAvailable = source["eventsAvailable"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class SchedulerAggregationRequest {
	    clusterId: string;
	    namespaces?: string[];
	    labelSelector?: string;
	    throttleMs?: number;
	    minPayload?: boolean;
	    include?: AggregationIncludeConfig;
	
	    static createFrom(source: any = {}) {
	        return new SchedulerAggregationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.clusterId = source["clusterId"];
	        this.namespaces = source["namespaces"];
	        this.labelSelector = source["labelSelector"];
	        this.throttleMs = source["throttleMs"];
	        this.minPayload = source["minPayload"];
	        this.include = this.convertValues(source["include"], AggregationIncludeConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SchedulerHealthSnapshot {
	    schedulerStatus?: Record<string, string>;
	    totalPending: number;
	    totalNodes: number;
	    totalGpus: number;
	    warnings?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SchedulerHealthSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schedulerStatus = source["schedulerStatus"];
	        this.totalPending = source["totalPending"];
	        this.totalNodes = source["totalNodes"];
	        this.totalGpus = source["totalGpus"];
	        this.warnings = source["warnings"];
	    }
	}
	export class SchedulerSnapshot {
	    clusterId: string;
	    // Go type: time
	    generatedAt: any;
	    seq: number;
	    nodeView: NodeViewSnapshot;
	    pendingView: PendingViewSnapshot;
	    draView: DRASnapshot;
	    migrationView: MigrationSnapshot;
	    health: SchedulerHealthSnapshot;
	
	    static createFrom(source: any = {}) {
	        return new SchedulerSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.clusterId = source["clusterId"];
	        this.generatedAt = this.convertValues(source["generatedAt"], null);
	        this.seq = source["seq"];
	        this.nodeView = this.convertValues(source["nodeView"], NodeViewSnapshot);
	        this.pendingView = this.convertValues(source["pendingView"], PendingViewSnapshot);
	        this.draView = this.convertValues(source["draView"], DRASnapshot);
	        this.migrationView = this.convertValues(source["migrationView"], MigrationSnapshot);
	        this.health = this.convertValues(source["health"], SchedulerHealthSnapshot);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace schema {
	
	export class GroupVersionResource {
	    Group: string;
	    Version: string;
	    Resource: string;
	
	    static createFrom(source: any = {}) {
	        return new GroupVersionResource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Group = source["Group"];
	        this.Version = source["Version"];
	        this.Resource = source["Resource"];
	    }
	}

}

