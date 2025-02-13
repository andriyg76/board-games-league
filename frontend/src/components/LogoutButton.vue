<template>
  <button
      @click="handleLogout"
      :disabled="loading"
      class="logout-button"
  >
    {{ loading ? 'Logging out...' : 'Logout' }}
  </button>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { AuthAPI } from '@/api/Auth';

const router = useRouter();
const loading = ref(false);

const handleLogout = async () => {
  loading.value = true;
  try {
    await AuthAPI.logout();
    await router.push('/login');
  } catch (error) {
    console.error('Logout failed:', error);
  } finally {
    loading.value = false;
  }
};
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