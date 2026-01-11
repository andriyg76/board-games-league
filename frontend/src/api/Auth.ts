import {User} from "@/api/UserApi";

export default {
    startLoginEntrypoint(provider: string) {
        const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
        let stateToken = "";
        for (let i = 0; i < 64; i++) {
            stateToken += letters.charAt(Math.floor(Math.random() * letters.length));
        }

        return `/api/auth/google?provider=${provider}&state=${stateToken}`;
    },

    async logout(): Promise<void> {
        const rotateToken = localStorage.getItem('rotateToken');
        try {
            const headers: HeadersInit = {};
            if (rotateToken) {
                headers['Authorization'] = `Bearer ${rotateToken}`;
            }

            const response = await fetch('/api/auth/logout', {
                method: 'POST',
                credentials: 'include',
                headers,
            });

            if (!response.ok) {
                throw new Error('Logout failed');
            }
        } finally {
            // Ensure rotate token is removed even if request fails
            localStorage.removeItem('rotateToken');
        }
    },
    async handleAuthCallback(params: string): Promise<User | null> {
        const response = await fetch(`/api/auth/google/callback?${params}`, {
            credentials: 'include',
            method: 'POST'
        });

        if (!response.ok) {
            throw new Error('Auth callback failed');
        }

        const data = await response.json() || {};
        
        // Store rotateToken in localStorage if provided
        if ((data as any).rotateToken) {
            localStorage.setItem('rotateToken', (data as any).rotateToken);
        }
        
        return data;
    }
}