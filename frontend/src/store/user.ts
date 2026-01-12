import { defineStore } from 'pinia';

interface User {
    external_ids?: string[];
    name: string;
    avatar?: string;
    alias: string;
    avatars?: string[];
    names?: string[];
    roles?: string[];
}

interface UserState {
    user: User;
    loggedIn: boolean;
}

export const useUserStore = defineStore('user', {
    state: (): UserState => ({
        user: {
            external_ids: [],
            name: '',
            alias: '',
        },
        loggedIn: false
    }),

    actions: {
        setUser(user: User) {
            this.user = user;
            this.loggedIn = (user.external_ids || []).length > 0;
        },

        clearUser() {
            this.user = {
                external_ids: [],
                name: '',
                alias: ''
            };
            this.loggedIn = false;
        }
    },

    getters: {
        isAuthenticated(): boolean {
            return this.loggedIn;
        },

        currentUser(): User {
            return this.user;
        },

        isSuperAdmin(): boolean {
            return this.user.roles?.includes('superadmin') ?? false;
        },

        hasRole(): (role: string) => boolean {
            return (role: string) => this.user.roles?.includes(role) ?? false;
        }
    }
});