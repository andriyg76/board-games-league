<template>
  <div class="mobile-accept">
    <n-spin v-if="loading" size="large" />

    <n-alert v-else-if="errorMessage" type="error">
      {{ errorMessage }}
    </n-alert>

    <div v-else-if="success" class="mobile-accept__result">
      <h2>{{ t('leagues.congratulations') }}</h2>
      <p>
        {{ t('leagues.joinedLeague') }}
        <strong>{{ joinedLeagueName }}</strong>
      </p>
      <n-button type="primary" @click="goToLeagueHome">{{ t('leagues.goToLeague') }}</n-button>
    </div>

    <div v-else-if="alreadyMember" class="mobile-accept__result">
      <h2>{{ t('leagues.alreadyMember') }}</h2>
      <p>{{ t('leagues.alreadyMemberDescription') }}</p>
      <n-button type="primary" @click="goToLeagueHome">{{ t('leagues.goToLeague') }}</n-button>
    </div>

    <div v-else class="mobile-accept__content">
      <h1 class="mobile-accept__title">{{ t('leagues.invitation') }}</h1>

      <div v-if="preview" class="mobile-accept__preview">
        <div class="mobile-accept__league">{{ preview.league_name }}</div>
        <div class="mobile-accept__meta">
          {{ t('leagues.invitedBy', { name: preview.inviter_alias }) }}
        </div>
        <div class="mobile-accept__meta">
          {{ t('leagues.youWillJoinAs', { alias: preview.player_alias }) }}
        </div>
      </div>

      <n-button type="primary" size="large" :loading="accepting" @click="handleAccept">
        Accept invitation
      </n-button>

      <div v-if="needsLogin" class="mobile-accept__login">
        <p>{{ t('leagues.loginToAcceptInvitation') }}</p>
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { NSpin, NAlert, NButton, NSpace, NIcon } from 'naive-ui';
import { LogoGoogle as LogoGoogleIcon, LogoDiscord as LogoDiscordIcon } from '@vicons/ionicons5';
import Auth from '@/api/Auth';
import { useLeagueStore } from '@/store/league';
import { useUserStore } from '@/store/user';
import type { InvitationPreview } from '@/api/LeagueApi';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const leagueStore = useLeagueStore();
const userStore = useUserStore();

const loading = ref(true);
const accepting = ref(false);
const needsLogin = ref(false);
const success = ref(false);
const alreadyMember = ref(false);
const errorMessage = ref<string | null>(null);
const preview = ref<InvitationPreview | null>(null);
const joinedLeagueName = ref('');
const alreadyMemberLeagueCode = ref<string | null>(null);

const autoAcceptTokenKey = 'invitation_auto_accept_token';

const token = computed(() => route.params.token as string | undefined);

const loadPreview = async () => {
  if (!token.value) {
    errorMessage.value = t('leagues.invalidInvitation');
    return;
  }
  preview.value = await leagueStore.previewInvitation(token.value);
  if (preview.value.status === 'expired') {
    errorMessage.value = t('leagues.invitationExpired');
  }
  if (preview.value.status === 'used') {
    errorMessage.value = t('leagues.invitationUsed');
  }
};

const acceptInvitation = async () => {
  if (!token.value) return;
  accepting.value = true;
  errorMessage.value = null;
  try {
    const result = await leagueStore.acceptInvitation(token.value);
    joinedLeagueName.value = result.league.name;
    success.value = true;
  } catch (error) {
    const errWithCode = error as Error & { leagueCode?: string };
    if (errWithCode.leagueCode) {
      alreadyMember.value = true;
      alreadyMemberLeagueCode.value = errWithCode.leagueCode;
      return;
    }
    errorMessage.value = error instanceof Error ? error.message : t('leagues.error');
  } finally {
    accepting.value = false;
  }
};

const handleAccept = async () => {
  if (!userStore.loggedIn) {
    needsLogin.value = true;
    if (token.value) {
      sessionStorage.setItem('invitation_return_url', route.fullPath);
      sessionStorage.setItem(autoAcceptTokenKey, token.value);
    }
    return;
  }
  await acceptInvitation();
};

const startLogin = (provider: string) => {
  if (token.value) {
    sessionStorage.setItem('invitation_return_url', route.fullPath);
  }
  const url = Auth.startLoginEntrypoint(provider);
  if (url) {
    window.location.href = url;
  }
};

const goToLeagueHome = async () => {
  if (alreadyMemberLeagueCode.value) {
    await leagueStore.setCurrentLeague(alreadyMemberLeagueCode.value);
  }
  router.push({ name: 'MobileLeagueHome' });
};

onMounted(async () => {
  loading.value = true;
  errorMessage.value = null;
  try {
    await loadPreview();
    if (errorMessage.value) return;

    if (userStore.loggedIn && token.value) {
      const pendingToken = sessionStorage.getItem(autoAcceptTokenKey);
      if (pendingToken === token.value) {
        sessionStorage.removeItem(autoAcceptTokenKey);
        await acceptInvitation();
      }
    }
  } catch (error) {
    console.error('Failed to load invitation:', error);
    errorMessage.value = t('leagues.error');
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.mobile-accept {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 24px;
}

.mobile-accept__title {
  margin: 0;
  font-size: 1.5rem;
}

.mobile-accept__preview {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 16px;
  border-radius: 12px;
  background: rgba(32, 128, 240, 0.08);
}

.mobile-accept__league {
  font-weight: 600;
}

.mobile-accept__meta {
  font-size: 0.9rem;
  opacity: 0.8;
}

.mobile-accept__login {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mobile-accept__result {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>
