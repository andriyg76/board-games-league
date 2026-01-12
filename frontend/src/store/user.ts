import { defineStore } from 'pinia';

interface User {
    external_ids?: string[];
    name: string;
    avatar?: string;
    alias: string;
    avatars?: string[];
    names?: string[];
    is_super_admin?: boolean;
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
        }
    }
});