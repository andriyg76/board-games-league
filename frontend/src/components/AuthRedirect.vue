<!-- AuthRedirect.vue -->
<template>
  <div class="redirect-loading">Redirecting...</div>
</template>

<script>
export default {
  name: 'AuthRedirect',
  created() {
    const
        url = this.$route.params.url,
        backurl = this.$route.params.backurl
    ;
    localStorage.setItem('auth_redirect', backurl  || this.$route.path);
    if (url) {
      window.location.href = url;
    } else {
      this.$router.push('/');
    }
  }
}
</script>