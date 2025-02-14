<template>
  <template v-if="loggedIn">
    <img :src="user.picture" v-if="user.picture" height="32" width="32" :alt="`${user.name} - ${user.email}`"/>
    <button class="logout-button" disabled v-if="loading">Logging out...</button>
    <button
        @click="handleLogout"
        :disabled="loading"
        class="logout-button"
        v-else
    >
      {{ loading ? 'Logging out...' : 'Logout' }}
    </button>
  </template>
  <button v-else class="logout-button" @click="router.push('/ui/user')">Login</button>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import Auth, { User } from '@/api/Auth';

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
    user.value = (await Auth.getUser()) || {};
  } catch (error) {
    console.error('Failed to get user:', error);
  }
});
</script>

<style scoped>
.logout-button {
  padding: 8px 16px;
  background-color: #f56565;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.logout-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>