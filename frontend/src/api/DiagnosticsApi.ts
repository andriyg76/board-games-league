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

export interface DiagnosticsResponse {
    server_info: {
        host_url: string;
        trusted_origins: string[];
    };
    build_info: BuildInfo;
    request_info: {
        ip_address: string;
        base_url: string;
        user_agent: string;
        origin: string;
        is_trusted: boolean;
        geo_info?: GeoIPInfo;
        resolution_info?: Record<string, string>;
    };
    runtime_info: RuntimeInfo;
    environment_vars: EnvVarInfo[];
}

export async function getFrontendBuildInfo(): Promise<BuildInfo> {
    try {
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
        const response = await fetch('/api/admin/diagnostics', {
            credentials: 'include',
        });

        if (!response.ok) {
            throw new Error('Failed to get diagnostics');
        }

        return await response.json();
    }
};
