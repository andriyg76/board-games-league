<template>
  <div>
    <google-login v-if="isLoggedIn"/>
    <template v-else>
      <h3>I am a {{ user.name }}</h3>
      <h4>Email {{ user.email }}</h4>
      <img v-if="user.picture" :src="user.picture" :alt="`${user.name} - ${user.email}`"/>
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import GoogleLogin from '@/components/GoogleLogin.vue';
import Auth, {User} from "@/api/Auth";

export default defineComponent({
  data() {
    return {
      user: {} as User,
    }
  },
  async mounted() {
    try {
      this.user = (await Auth.getUser()) || {}
    } catch (e) {
      console.error("Error getting user: ", e)
    }
  },
  components: {
    GoogleLogin
  }
});
</script>