<template>
  <div class="callback-loading">Completing authentication...</div>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';
import Auth from '@/api/Auth';
import userStore from '@/store/user';

const router = useRouter();

try {
  // Get all query parameters from current URL
  const queryParams = new URLSearchParams(window.location.search);

  Auth.handleGoogleCallback(queryParams.toString())
      .then((user) => {
        if (user) {
          userStore.setUser(user);
          console.log("User authenticated: ", user);
        } else {
          console.log("User is not authenticated");
        }
      })
      .finally(() => {
        const redirectPath = localStorage.getItem('auth_redirect') || '/';
        localStorage.removeItem('auth_redirect'); // Clean up
        console.info("Redirecting to: ", redirectPath);
        router.push(redirectPath);
      });
} catch (error) {
  console.error('Auth callback failed:', error);
  router.push('/ui/user');
}
</script>