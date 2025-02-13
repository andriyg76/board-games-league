<template>
  <template v-if="loggedIn">
    <button class="logout-button" disabled v-if="loading">Logging out...</button>
    <button
        @click="handleLogout"
        :disabled="loading"
        class="logout-button"
        v-else
    >
      {{ loading ? 'Logging out...' : 'Logout' }}
    </button>
    <img :src="user.picture" v-if="user.picture" height="32" width="32" alt="{{ user.name }} - {{ user.email }}"/>
  </template>
  <button v-else class="logout-button" @click="router.push('/ui/user')">Login</button>
</template>

<script lang="ts">
import {defineComponent} from 'vue';
import { useRouter } from 'vue-router';
import Auth, {User} from '@/api/Auth';

export default defineComponent({
  data() {
    return {
      user: {} as User,
      loading: false,
      router: useRouter()
    }
  },
  async mounted() {
    try {
      this.user = (await Auth.getUser()) || {}
    } catch(error) {
      console.error('Failed to get user:', error);
    }
  },
  computed: {
    loggedIn() {
      return !!this.user.email;
    }
  },
  methods: {
    async handleLogout(){
      this.loading = true;
      try {
        await Auth.logout();
        await this.router.push('/login');
      } catch (error) {
        console.error('Logout failed:', error);
      } finally {
        this.loading = false;
      }
    }
  }
})

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