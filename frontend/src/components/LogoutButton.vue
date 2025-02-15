<template>
  <template v-if="userStore.state.loggedIn">
    <v-btn color="primary"
           @click="handleLogout"
           :disabled="loading"
           class="logout-button"
    >
      {{ loading ? 'Logging out...' : 'Logout' }}
      <v-img  :src="userStore.state.user.picture" v-if="userStore.state.user.picture" height="32" width="32" :alt="`${userStore.state.user.name} - ${userStore.state.user.email}`"/>
    </v-btn>
  </template>
  <v-btn color="primary" v-else class="logout-button" @click="router.push('/ui/user')">Login</v-btn>
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