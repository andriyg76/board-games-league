<template>
  <div>
    <google-login/>
    <h3>I am a {{ user }}</h3>
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