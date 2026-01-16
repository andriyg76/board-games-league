<template>
  <template v-if="userStore.loggedIn">
    <n-button
        type="primary"
        @click="handleLogout"
        :loading="loading"
        class="logout-button"
        data-testid="auth-logout-button"
    >
      <img v-if="userStore.user.avatar" :src="userStore.user.avatar" height="32" width="32" :alt="`${userStore.user.name} - ${userStore.user.alias}`" style="margin-right: 8px; border-radius: 50%; object-fit: cover;" />
      {{ loading ? t('auth.loggingOut') : t('auth.logout') }}
    </n-button>
  </template>
  <n-dropdown v-else :options="loginOptions" trigger="click" @select="startLogin">
    <n-button type="primary" data-testid="auth-login-button">
      {{ t('auth.login') }}
    </n-button>
  </n-dropdown>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import { NButton, NDropdown } from 'naive-ui';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import Auth from '@/api/Auth';
import { useUserStore } from '@/store/user';
import UserApi from "@/api/UserApi";

const userStore = useUserStore();
const { t } = useI18n();
const loading = ref(false);
const router = useRouter();

const loginOptions = computed(() => [
  {
    label: t('auth.loginWithGoogle'),
    key: 'google',
  },
  {
    label: t('auth.loginWithDiscord'),
    key: 'discord',
  },
]);

const handleLogout = async () => {
  loading.value = true;
  try {
    await Auth.logout();
    userStore.clearUser();
    await router.push('/login');
  } catch (error) {
    console.error('Logout failed:', error);
  } finally {
    loading.value = false;
  }
};

const startLogin = async (provider: string) => {
  try {
    let url = Auth.startLoginEntrypoint(provider);
    console.info("Redirecting to: ", url)
    loading.value = true;
    // Store the current route to redirect back after auth
    const currentRoute = router.currentRoute.value;
    localStorage.setItem('auth_redirect', currentRoute.fullPath);
    if (url) {
      window.location.href = url;
    }
  } catch (e) {
    console.error("error login start", e);
  } finally {
    loading.value = false;
  }
};

onMounted(async () => {
  try {
    const user = await UserApi.getUser();
    if (user) {
      userStore.setUser(user);
    }
  } catch (e) {
    console.error("Error getting user: ", e);
  }
});
</script>
