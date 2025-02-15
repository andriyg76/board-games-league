export interface User {
    email?: string | null;
    name?: string | null;
    picture?: string | null;
}



export  default {
    get googleLoginEntrypoint() {
        const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
        let stateToken = "";
        for (let i = 0; i < 64; i++) {
            stateToken += letters.charAt(Math.floor(Math.random() * letters.length));
        }

        return `/api/auth/google?provider=google&state=${stateToken}`;
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
            return null;
        }

        if (!response.ok) {
            throw new Error('Failed to get user');
        }

        return await response.json();
    },
    async handleGoogleCallback(params: string) : Promise<User | null> {
        // Forward these parameters to your backend
        const response = await fetch(`/api/auth/google/callback?${params}`,  {
            credentials: 'include',
            method: 'POST'
        });

        if (!response.ok) {
            throw new Error('Auth callback failed');
        }

        return await response.json() || {}
    }
}