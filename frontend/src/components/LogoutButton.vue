<template>
  <template v-if="userStore.loggedIn">
    <v-btn color="primary"
           @click="handleLogout"
           :disabled="loading"
           class="logout-button"
    >
      <v-img  :src="userStore.user.avatar" v-if="userStore.user.avatar" height="32" width="32" :alt="`${userStore.user.name} - ${userStore.user.alias}`"/>
      {{ loading ? 'Logging out...' : 'Logout' }}
    </v-btn>
  </template>
  <v-menu v-else>
    <template v-slot:activator="{ props }">
      <v-btn color="primary" v-bind="props">
        Login
      </v-btn>
    </template>
    <v-list>
      <v-list-item @click="startLogin('google')">
        <v-list-item-title>Login with Google</v-list-item-title>
      </v-list-item>
      <v-list-item @click="startLogin('discord')">
        <v-list-item-title>Login with Discord</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue';
import { useRouter } from 'vue-router';
import Auth from '@/api/Auth';
import { useUserStore } from '@/store/user';
const userStore = useUserStore();
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

const startLogin = async (provider: string) => {
  try {
    let url = Auth.startLoginEntrypoint(provider);
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