import { apiFetch } from './apiClient';

export interface GeoIPInfo {
    country?: string;
    country_code?: string;
    region?: string;
    region_name?: string;
    city?: string;
    timezone?: string;
    isp?: string;
    ip?: string;
}

export interface BuildInfo {
    version: string;
    commit: string;
    branch: string;
    date: string;
}

export interface ServerInfo {
    host_url: string;
    trusted_origins: string[];
}

export interface RequestInfo {
    ip_address: string;
    base_url: string;
    user_agent: string;
    origin: string;
    is_trusted: boolean;
    geo_info?: GeoIPInfo;
    resolution_info?: Record<string, string>;
}

export interface MemoryInfo {
    alloc_bytes: number;
    total_alloc_bytes: number;
    sys_bytes: number;
    heap_alloc_bytes: number;
    heap_sys_bytes: number;
    heap_inuse_bytes: number;
    num_gc: number;
}

export interface RuntimeInfo {
    go_version: string;
    goos: string;
    goarch: string;
    num_cpu: number;
    num_goroutine: number;
    uptime: string;
    uptime_seconds: number;
    start_time: string;
    memory: MemoryInfo;
}

export interface EnvVarInfo {
    name: string;
    value: string;
    masked: boolean;
}

export interface CacheStatsInfo {
    name: string;
    current_size: number;
    max_size: number;
    expired_count: number;
    ttl: string;
    ttl_seconds: number;
    usage_percent: number;
}

export interface DiagnosticsRequestResponse {
    server_info: ServerInfo;
    request_info: RequestInfo;
}

export interface DiagnosticsSystemResponse {
    runtime_info: RuntimeInfo;
    environment_vars: EnvVarInfo[];
    cache_stats?: CacheStatsInfo[];
}

export interface DiagnosticsBuildResponse {
    build_info: BuildInfo;
}

export interface DiagnosticsResponse {
    server_info: ServerInfo;
    build_info: BuildInfo;
    request_info: RequestInfo;
    runtime_info: RuntimeInfo;
    environment_vars: EnvVarInfo[];
    cache_stats?: CacheStatsInfo[];
}

export async function getFrontendBuildInfo(): Promise<BuildInfo> {
    try {
        // Note: version.json is a static file that doesn't require auth
        const response = await fetch('/version.json', {
            credentials: 'include',
        });
        if (response.ok) {
            return await response.json();
        }
    } catch (e) {
        console.error("Failed to load frontend version info:", e);
    }
    return {
        version: "unknown",
        commit: "unknown",
        branch: "unknown",
        date: "unknown",
    };
}

export default {
    async getDiagnostics(): Promise<DiagnosticsResponse> {
        const response = await apiFetch('/api/admin/diagnostics');

        if (!response.ok) {
            throw new Error('Failed to get diagnostics');
        }

        return await response.json();
    },
    async getRequestDiagnostics(): Promise<DiagnosticsRequestResponse> {
        const response = await apiFetch('/api/admin/diagnostics?sections=request');

        if (!response.ok) {
            throw new Error('Failed to get request diagnostics');
        }

        return await response.json();
    },
    async getSystemDiagnostics(): Promise<DiagnosticsSystemResponse> {
        const response = await apiFetch('/api/admin/diagnostics?sections=system');

        if (!response.ok) {
            throw new Error('Failed to get system diagnostics');
        }

        return await response.json();
    },
    async getBuildDiagnostics(): Promise<DiagnosticsBuildResponse> {
        const response = await apiFetch('/api/admin/diagnostics?sections=build');

        if (!response.ok) {
            throw new Error('Failed to get build diagnostics');
        }

        return await response.json();
    },
};
