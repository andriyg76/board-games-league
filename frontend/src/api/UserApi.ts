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
    async checkAlias(alias: string | null): Promise<{ isUnique: boolean }> {
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
    },
    async adminCreateUser(email: string): Promise<void> {
        try {
            const response = await fetch('/api/admin/user/create', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email }),
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
    }
}