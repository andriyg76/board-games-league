/**
 * Centralized API client with automatic token refresh on 401
 * 
 * This module provides a fetch wrapper that:
 * 1. Always includes credentials (cookies) with requests
 * 2. Automatically attempts to refresh the auth token on 401 responses
 * 3. Retries the original request after successful refresh
 * 4. Clears auth state if refresh fails (session expired)
 */

// Track ongoing refresh to prevent multiple simultaneous refresh attempts
let refreshPromise: Promise<boolean> | null = null;

/**
 * Attempt to refresh the auth token using the rotate token from localStorage
 * Returns true if refresh succeeded, false otherwise
 */
async function refreshAuthToken(): Promise<boolean> {
    const rotateToken = localStorage.getItem('rotateToken');
    if (!rotateToken) {
        return false;
    }

    try {
        const response = await fetch('/api/auth/refresh', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Authorization': `Bearer ${rotateToken}`,
                'Content-Type': 'application/json',
            },
        });

        if (!response.ok) {
            // Refresh failed - token is invalid or expired
            localStorage.removeItem('rotateToken');
            return false;
        }

        const data = await response.json();
        
        // If a new rotate token was provided, update it
        if (data.rotateToken) {
            localStorage.setItem('rotateToken', data.rotateToken);
        }

        return true;
    } catch (error) {
        console.error('Token refresh failed:', error);
        localStorage.removeItem('rotateToken');
        return false;
    }
}

/**
 * Wrapper around fetch that automatically:
 * - Includes credentials
 * - Handles 401 by refreshing token and retrying
 */
export async function apiFetch(
    url: string,
    options: RequestInit = {}
): Promise<Response> {
    // Always include credentials for auth cookies
    const fetchOptions: RequestInit = {
        ...options,
        credentials: 'include',
    };

    let response = await fetch(url, fetchOptions);

    // If unauthorized, try to refresh and retry
    if (response.status === 401) {
        // Deduplicate concurrent refresh attempts
        if (!refreshPromise) {
            refreshPromise = refreshAuthToken().finally(() => {
                refreshPromise = null;
            });
        }

        const refreshed = await refreshPromise;
        
        if (refreshed) {
            // Retry the original request
            response = await fetch(url, fetchOptions);
        }
        // If refresh failed, return the original 401 response
        // The caller can handle it (e.g., redirect to login)
    }

    return response;
}

/**
 * Helper for JSON API requests
 */
export async function apiJson<T>(
    url: string,
    options: RequestInit = {}
): Promise<T> {
    const response = await apiFetch(url, options);
    
    if (!response.ok) {
        throw new Error(`API request failed: ${response.status} ${response.statusText}`);
    }
    
    return response.json();
}

/**
 * Helper for POST/PUT JSON requests
 */
export async function apiJsonPost<T>(
    url: string,
    body: unknown,
    method: 'POST' | 'PUT' | 'PATCH' = 'POST'
): Promise<T> {
    return apiJson<T>(url, {
        method,
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
    });
}

export default {
    fetch: apiFetch,
    json: apiJson,
    post: apiJsonPost,
};

