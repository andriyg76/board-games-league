<template>
  <template v-if="loggedIn">
    <v-btn color="primary"
        @click="handleLogout"
        :disabled="loading"
        class="logout-button"
    >
      {{ loading ? 'Logging out...' : 'Logout' }}
      <v-img  :src="user.picture" v-if="user.picture" height="32" width="32" :alt="`${user.name} - ${user.email}`"/>
    </v-btn>
  </template>
  <v-btn color="primary" v-else class="logout-button" @click="router.push('/ui/user')">Login</v-btn>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import Auth from '@/api/Auth';
import UserApi, { User } from "@/api/UserApi";

const user = ref<User>({} as User);
const loading = ref(false);
const router = useRouter();

const loggedIn = computed(() => !!user.value.email);

const handleLogout = async () => {
  loading.value = true;
  try {
    await Auth.logout();
    await router.push('/login');
  } catch (error) {
    console.error('Logout failed:', error);
  } finally {
    loading.value = false;
  }
};

onMounted(async () => {
  try {
    user.value = (await UserApi.getUser()) || {};
  } catch (error) {
    console.error('Failed to get user:', error);
  }
});
</script>

