<!-- AuthRedirect.vue -->
<template>
  <div class="redirect-loading">Redirecting...</div>
</template>

<script>
export default {
  name: 'AuthRedirect',
  created() {
    const url = this.$route.params.url;
    if (url) {
      window.location.href = url;
    } else {
      this.$router.push('/');
    }
  }
}
</script>