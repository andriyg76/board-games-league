export interface User {
    email?: string | null;
    name?: string | null;
    picture?: string | null;
}


export  default {
    get googleLoginEntrypoint() {
        return `/api/auth/google?random=${Math.random()}`
    },

    async logout(): Promise<void> {
        const response = await fetch('/api/auth/logout', {
            method: 'POST',
            credentials: 'include',
        });

        if (!response.ok) {
            throw new Error('Logout failed');
        }
    },
    async getUser(): Promise<User | null> {
        const response = await fetch('/api/user', {
            credentials: 'include',
        });

        if (response.status == 401) {
            return Promise.resolve(null);
        }

        if (!response.ok) {
            throw new Error('Failed to get user');
        }

        return response.json();
    },
    async handleGoogleCallback(params: string) {
        // Forward these parameters to your backend
        const response = await fetch(`/api/auth/google/callback?${params}`,  {
            credentials: 'include',
            method: 'POST'
        });

        if (!response.ok) {
            throw new Error('Auth callback failed');
        }
    }
}