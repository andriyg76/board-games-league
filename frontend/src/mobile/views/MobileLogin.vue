<template>
  <div class="mobile-login">
    <div class="mobile-login__content">
      <h1 class="mobile-login__title">{{ t('auth.login') }}</h1>
      <p class="mobile-login__subtitle">Login to continue</p>
      <n-space vertical size="large">
        <n-button type="primary" @click="startLogin('google')">
          <template #icon>
            <n-icon><LogoGoogleIcon /></n-icon>
          </template>
          {{ t('auth.loginWithGoogle') }}
        </n-button>
        <n-button type="primary" color="#5865F2" @click="startLogin('discord')">
          <template #icon>
            <n-icon><LogoDiscordIcon /></n-icon>
          </template>
          {{ t('auth.loginWithDiscord') }}
        </n-button>
      </n-space>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { NButton, NSpace, NIcon } from 'naive-ui';
import { LogoGoogle as LogoGoogleIcon, LogoDiscord as LogoDiscordIcon } from '@vicons/ionicons5';
import Auth from '@/api/Auth';
import { useUserStore } from '@/store/user';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

const startLogin = (provider: string) => {
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/m';
  localStorage.setItem('auth_redirect', redirect);
  const url = Auth.startLoginEntrypoint(provider);
  if (url) {
    window.location.href = url;
  }
};

onMounted(() => {
  if (userStore.loggedIn) {
    router.replace(typeof route.query.redirect === 'string' ? route.query.redirect : '/m');
  }
});
</script>

<style scoped>
.mobile-login {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  padding: 32px 24px;
  text-align: center;
}

.mobile-login__content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.mobile-login__title {
  font-size: 1.5rem;
  margin: 0;
}

.mobile-login__subtitle {
  opacity: 0.7;
  margin: 0;
}
</style>
