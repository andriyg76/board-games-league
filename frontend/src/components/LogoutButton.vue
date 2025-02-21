<template>
  <template v-if="userStore.state.loggedIn">
    <v-btn color="primary"
           @click="handleLogout"
           :disabled="loading"
           class="logout-button"
    >
      <v-img  :src="userStore.state.user.avatar" v-if="userStore.state.user.avatar" height="32" width="32" :alt="`${userStore.state.user.name} - ${userStore.state.user.alias}`"/>
      {{ loading ? 'Logging out...' : 'Logout' }}
    </v-btn>
  </template>
  <v-btn color="primary" v-else @click="startLogin">Login</v-btn>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue';
import { useRouter } from 'vue-router';
import Auth from '@/api/Auth';
import userStore from '@/store/user';
import UserApi from "@/api/UserApi";

const loading = ref(false);
const router = useRouter();

const handleLogout = async () => {
  loading.value = true;
  try {
    await Auth.logout();
    userStore.clearUser();
    await router.push('/login');
  } catch (error) {
    console.error('Logout failed:', error);
  } finally {
    loading.value = false;
  }
};

const startLogin = async () => {
  try {
    let url = Auth.googleLoginEntrypoint;
    console.info("Redirecting to: ", url)
    loading.value = true;
    // Store the current route to redirect back after auth
    const currentRoute = router.currentRoute.value;
    localStorage.setItem('auth_redirect', currentRoute.fullPath);
    if (url) {
      window.location.href = url;
    }
  } catch (e) {
    console.error("error login start", e);
  } finally {
    loading.value = false;
  }
};

onMounted(async () => {
  try {
    const user = await UserApi.getUser();
    if (user) {
      userStore.setUser(user);
    }
  } catch (e) {
    console.error("Error getting user: ", e);
  }
});
</script>