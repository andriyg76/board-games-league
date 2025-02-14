<template>
  <div class="callback-loading">Completing authentication...</div>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';
import Auth from '@/api/Auth';

const router = useRouter();

try {
  // Get all query parameters from current URL
  const queryParams = new URLSearchParams(window.location.search);

  await Auth.handleGoogleCallback(queryParams.toString());

  const redirectPath = localStorage.getItem('auth_redirect') || '/';
  localStorage.removeItem('auth_redirect'); // Clean up

  router.push(redirectPath);
} catch (error) {
  console.error('Auth callback failed:', error);
  router.push('/ui/user');
}
</script>