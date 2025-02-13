<template>
  <div class="callback-loading">Completing authentication...</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
export default defineComponent( {
  name: 'AuthCallback',
  async created() {
    try {
      // Get all query parameters from current URL
      const queryParams = new URLSearchParams(window.location.search);

      // Forward these parameters to your backend
      const response = await fetch(`/api/auth/google/callback?${queryParams.toString()}`);

      if (!response.ok) {
        throw new Error('Auth callback failed');
      }

      const redirectPath = localStorage.getItem('auth_redirect') || '/dashboard';
      localStorage.removeItem('auth_redirect'); // Clean up

      this.$router.push(redirectPath);
    } catch (error) {
      console.error('Auth callback failed:', error);
      this.$router.push('/login');
    }
  }
})
</script>