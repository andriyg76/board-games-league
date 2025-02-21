import { reactive } from 'vue';

interface User {
    externalIDs: string[];
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
    state.loggedIn = (user.externalIDs || []).length > 0;
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