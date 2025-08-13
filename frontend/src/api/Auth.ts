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
        const response = await fetch('/api/auth/logout', {
            method: 'POST',
            credentials: 'include',
        });

        if (!response.ok) {
            throw new Error('Logout failed');
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

        return await response.json() || {};
    }
}