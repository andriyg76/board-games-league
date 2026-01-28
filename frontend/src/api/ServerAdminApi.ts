export interface DebugLoggingState {
    enabled: boolean;
    expires_at?: string;
}

export interface EnableDebugRequest {
    duration_minutes: number;
}

const TOKEN_STORAGE_KEY = 'admin_api_token';

export function getStoredToken(): string | null {
    return localStorage.getItem(TOKEN_STORAGE_KEY);
}

export function setStoredToken(token: string): void {
    localStorage.setItem(TOKEN_STORAGE_KEY, token);
}

export function clearStoredToken(): void {
    localStorage.removeItem(TOKEN_STORAGE_KEY);
}

async function apiRequest<T>(
    endpoint: string,
    options: RequestInit = {},
    isDownload = false
): Promise<T> {
    const token = getStoredToken();
    if (!token) {
        throw new Error('No admin token provided');
    }

    const headers: Record<string, string> = {
        'Authorization': `Bearer ${token}`,
    };

    // Only add Content-Type for non-GET requests that have a body
    if (options.method && options.method !== 'GET' && options.body) {
        headers['Content-Type'] = 'application/json';
    }

    // Merge with any provided headers
    if (options.headers) {
        Object.assign(headers, options.headers);
    }

    const response = await fetch(`/api/admin/server${endpoint}`, {
        ...options,
        headers,
    });

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`API error: ${response.status} - ${errorText}`);
    }

    // For download endpoints, return blob
    if (isDownload) {
        return response.blob() as unknown as T;
    }

    return response.json();
}

export default {
    async enableDebugLogging(durationMinutes: number): Promise<DebugLoggingState> {
        return apiRequest<DebugLoggingState>('/debug/enable', {
            method: 'POST',
            body: JSON.stringify({ duration_minutes: durationMinutes }),
        });
    },

    async disableDebugLogging(): Promise<DebugLoggingState> {
        return apiRequest<DebugLoggingState>('/debug/disable', {
            method: 'POST',
        });
    },

    async getDebugLoggingStatus(): Promise<DebugLoggingState> {
        return apiRequest<DebugLoggingState>('/debug/status');
    },

    async downloadLogsByPeriod(durationMinutes: number, files: string[]): Promise<Blob> {
        const filesParam = files.join(',');
        return apiRequest<Blob>(`/logs/download?duration_minutes=${durationMinutes}&files=${encodeURIComponent(filesParam)}`, {
            method: 'GET',
        }, true);
    },

    async downloadFullLogs(files: string[]): Promise<Blob> {
        const filesParam = files.join(',');
        return apiRequest<Blob>(`/logs/download-full?files=${encodeURIComponent(filesParam)}`, {
            method: 'GET',
        }, true);
    },
};
