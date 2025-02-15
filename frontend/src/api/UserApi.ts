export interface User {
    email?: string;
    name?: string;
    picture?: string;
    alias?: string;
}

export default {
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
    async checkAlias(alias: string): Promise<{ isUnique: boolean }> {
        const response = await fetch(`/api/user/alias/exist?alias=${alias}`, {
            credentials: 'include',
            method: "POST"
        });

        if (!response.ok) {
            throw new Error('Failed to check alias uniqunes');
        }

        return await response.json();
    },
    async updateUser(user: User): Promise<void> {
        const response = await fetch('/api/user/update', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify(user),
        });

        if (!response.ok) {
            throw new Error('Failed to update user');
        }
    }
}