<h3>I am user</h3>

<template>
  <div>
    <h3>I am a {{ user }}</h3>

    <google-sign-in url="url"/>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import GoogleSignIn from '@/components/GoogleSignin.vue';

export default defineComponent({
  data() {
    return {
      user: 'user',
      url: null,
    }
  },
  mounted() {
    function handleResponse(response: Response, target: { url: string | null; user: string } ) {
      if (response.status === 200) {
        response.json().then(data => target.user = data)
      } if (response.status === 401) {
        target.url = response.headers.get('X-Auth-URL');
      }
    }
    fetch('/api/user').then(r => handleResponse(r, this));
  },
  components: {
    GoogleSignIn
  }
});
</script>