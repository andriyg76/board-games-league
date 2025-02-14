<template>
  <div class="redirect-loading">Redirecting...</div>
</template>

<script lang="ts" setup>
import {useRoute, useRouter} from 'vue-router';

const route = useRoute();
const router = useRouter();

const url = route.params.url as string | undefined;
const backurl = route.params.backurl as string | undefined;

localStorage.setItem('auth_redirect', backurl || route.path);

if (url) {
  window.location.href = url;
} else {
  router.push('/');
}
</script>