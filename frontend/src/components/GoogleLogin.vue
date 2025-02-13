<!-- GoogleSignIn.vue -->
<template>
  <div class="google-signin">
    <div v-if="loading" class="loading">
      Loading...
    </div>
    <button
        v-else
        @click="handleSignIn"
        class="google-btn"
        :disabled="loading"
    >
      <span class="google-icon">
        <svg viewBox="0 0 24 24">
          <path fill="currentColor" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
          <path fill="currentColor" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
          <path fill="currentColor" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
          <path fill="currentColor" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
        </svg>
      </span>
      Login in with Google
    </button>
    <logout-button/>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import LogoutButton from "@/components/LogoutButton.vue";

export default defineComponent({
  name: 'GoogleLogin',
  components: {LogoutButton},
  props: {
    url: null,
    backUrl: null,
  },
  data() {
    return {
      loading: false,
      theUrl: this.url,
    }
  },
  methods: {
    async handleSignIn() {
      if (!this.theUrl) {
        this.loading = true;
        try {
          const response = await fetch('/api/user');
          if (response.status === 401) {
            const authUrl = response.headers.get('X-Auth-URL');
            if (authUrl) {
              this.theUrl = authUrl;
            } else {
              console.error('No auth URL provided');
            }
          }
        } catch (error) {
          console.error('Sign in failed:', error);
        } finally {
          this.loading = false;
        }
      }

      if (this.theUrl) {
        // Store the current route to redirect back after auth
        const currentRoute = this.$router.currentRoute.value;
        localStorage.setItem('auth_redirect', this.backUrl || currentRoute.fullPath);

        this.$router.push({
          name: 'auth-redirect',
          params: { url: this.theUrl },
        });
      }
    },
  }
});
</script>

<style scoped>
.google-signin {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100px;
}

.google-btn {
  display: flex;
  align-items: center;
  padding: 12px 24px;
  background-color: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 16px;
  font-weight: 500;
  color: #757575;
  cursor: pointer;
  transition: background-color 0.2s;
}

.google-btn:hover {
  background-color: #f8f8f8;
}

.google-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.google-icon {
  width: 20px;
  height: 20px;
  margin-right: 12px;
}

.loading {
  color: #666;
}
</style>