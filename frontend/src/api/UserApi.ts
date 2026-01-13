import { apiFetch } from './apiClient';

export interface User {
    external_ids?: string[];
    name?: string;
    avatar?: string;
    alias?: string;
    aliases?: string[];
    names?: string[];
}

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

export interface SessionInfo {
    id: string;
    ip_address: string;
    user_agent: string;
    created_at: string;
    updated_at: string;
    last_rotation_at: string;
    expires_at: string;
    is_current: boolean;
    geo_info?: GeoIPInfo;
}

export default {
    async getUser(): Promise<User | null> {
        const response = await apiFetch('/api/user');

        if (response.status == 401) {
            return null;
        }

        if (!response.ok) {
            throw new Error('Failed to get user');
        }

        return await response.json();
    },
    async checkAlias(alias: string | null): Promise<{ isUnique: boolean }> {
        const response = await apiFetch(`/api/user/alias/exist?alias=${alias}`, {
            method: "POST"
        });

        if (!response.ok) {
            throw new Error('Failed to check alias uniqunes');
        }

        return await response.json();
    },
    async updateUser(user: User): Promise<void> {
        const response = await apiFetch('/api/user/update', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(user),
        });

        if (!response.ok) {
            throw new Error('Failed to update user');
        }
    },
    async adminCreateUser(external_ids: string[]): Promise<void> {
        try {
            const response = await apiFetch('/api/admin/user/create', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ external_ids }),
            });

            if (response.status === 201) {
                console.log('User created successfully');
            } else {
                const errorData = await response.text();
                console.error('Failed to create user:', errorData);
            }
        } catch (error) {
            console.error('Failed to create user:', error);
            throw new Error('Error creating user:' + error);
        }
    },
    async getUserSessions(currentRotateToken?: string): Promise<SessionInfo[]> {
        const headers: HeadersInit = {};
        if (currentRotateToken) {
            headers['Authorization'] = `Bearer ${currentRotateToken}`;
        }

        const response = await apiFetch('/api/user/sessions', {
            headers,
        });

        if (!response.ok) {
            throw new Error('Failed to get user sessions');
        }

        return await response.json();
    }
}
