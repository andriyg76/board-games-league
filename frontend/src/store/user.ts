import { reactive } from 'vue';

interface User {
    name: string;
    email: string;
    picture?: string;
}

const state = reactive({
    user: {} as User,
    loggedIn: false,
});

const setUser = (user: User) => {
    state.user = user;
    state.loggedIn = !!user.email;
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