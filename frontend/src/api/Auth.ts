export interface User {
    email?: string | null;
    name?: string | null;
    picture?: string | null;
}

export  default {
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
    }
}