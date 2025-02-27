import { reactive } from 'vue';

interface User {
    external_ids: string[];
    name: string;
    avatar?: string;
    alias: string;
}

const state = reactive({
    user: {} as User,
    loggedIn: false,
});

const setUser = (user: User) => {
    state.user = user;
    state.loggedIn = (user.external_ids || []).length > 0;
};

const clearUser = () => {
    state.user = {} as User;
    state.loggedIn = false;
};

export default {
    state,
    setUser,
    clearUser,
};