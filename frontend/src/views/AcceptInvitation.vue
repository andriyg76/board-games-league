<template>
  <div>
    <n-grid :cols="24" :x-gap="16" justify="center">
      <n-gi :span="24" :responsive="{ m: 12 }">
        <n-card>
          <template #header>
            <div style="text-align: center; padding: 24px; background: #2080f0; color: white;">
              <n-icon size="32" style="margin-right: 8px; vertical-align: middle;">
                <MailOpenIcon />
              </n-icon>
              <span style="font-size: 1.25rem; font-weight: 500;">{{ t('leagues.invitation') }}</span>
            </div>
          </template>

          <!-- Loading State -->
          <div v-if="loading" style="text-align: center; padding: 64px 0;">
            <n-spin size="large" style="margin-bottom: 16px;" />
            <div style="font-size: 1.125rem;">{{ loadingMessage }}</div>
          </div>

          <!-- Login Required State (with preview) -->
          <div v-else-if="needsLogin" style="text-align: center; padding: 64px 0;">
            <n-icon size="80" color="#2080f0" style="margin-bottom: 16px; display: block;">
              <LogInIcon />
            </n-icon>

            <!-- Show invitation preview if available -->
            <template v-if="preview">
              <div style="font-size: 1.125rem; font-weight: 500; margin-bottom: 8px;">{{ preview.league_name }}</div>
              <div style="margin-bottom: 8px;">
                {{ t('leagues.invitedBy', { name: preview.inviter_alias }) }}
              </div>
              <div style="opacity: 0.7; margin-bottom: 16px;">
                {{ t('leagues.youWillJoinAs', { alias: preview.player_alias }) }}
              </div>
            </template>

            <div style="font-size: 1.125rem; font-weight: 500; margin-bottom: 8px;">{{ t('leagues.pleaseLoginFirst') }}</div>
            <div style="margin-bottom: 16px;">
              {{ t('leagues.loginToAcceptInvitation') }}
            </div>

            <n-space justify="center" :wrap="true">
              <n-button type="primary" tag="a" href="/oauth/login/google">
                <template #icon>
                  <n-icon><LogoGoogleIcon /></n-icon>
                </template>
                {{ t('auth.loginWithGoogle') }}
              </n-button>
              <n-button type="primary" tag="a" href="/oauth/login/discord" style="background: #5865F2;">
                <template #icon>
                  <n-icon><LogoDiscordIcon /></n-icon>
                </template>
                {{ t('auth.loginWithDiscord') }}
              </n-button>
            </n-space>
          </div>

          <!-- Already Member State -->
          <div v-else-if="alreadyMember" style="text-align: center; padding: 64px 0;">
            <n-icon size="80" color="#2080f0" style="margin-bottom: 16px; display: block;">
              <PersonCheckIcon />
            </n-icon>
            <div style="font-size: 1.125rem; font-weight: 500; margin-bottom: 8px;">{{ t('leagues.alreadyMember') }}</div>
            <div style="margin-bottom: 16px;">
              {{ t('leagues.alreadyMemberDescription') }}
            </div>

            <n-space justify="center">
              <n-button
                v-if="alreadyMemberLeagueCode"
                type="primary"
                @click="goToLeagueByCode(alreadyMemberLeagueCode)"
              >
                <template #icon>
                  <n-icon><ArrowForwardIcon /></n-icon>
                </template>
                {{ t('leagues.goToLeague') }}
              </n-button>
              <n-button @click="goToHome">
                {{ t('leagues.goToHome') }}
              </n-button>
            </n-space>
          </div>

          <!-- Error State -->
          <div v-else-if="error" style="text-align: center; padding: 64px 0;">
            <n-icon size="80" color="#d03050" style="margin-bottom: 16px; display: block;">
              <AlertCircleIcon />
            </n-icon>
            <div style="font-size: 1.125rem; font-weight: 500; margin-bottom: 8px;">{{ t('leagues.error') }}</div>
            <n-alert type="error" style="margin: 16px 0;">
              {{ error }}
            </n-alert>

            <n-button type="primary" @click="goToHome" style="margin-top: 16px;">
              {{ t('leagues.goToHome') }}
            </n-button>
          </div>

          <!-- Success State -->
          <div v-else-if="success && league" style="text-align: center; padding: 64px 0;">
            <n-icon size="80" color="#18a058" style="margin-bottom: 16px; display: block;">
              <CheckCircleIcon />
            </n-icon>
            <div style="font-size: 1.125rem; font-weight: 500; margin-bottom: 8px;">{{ t('leagues.congratulations') }}</div>
            <div style="margin-bottom: 16px;">
              {{ t('leagues.joinedLeague') }} <strong>{{ league.name }}</strong>
            </div>

            <n-divider style="margin: 16px 0;" />

            <n-card style="margin-bottom: 16px; background: rgba(32, 128, 240, 0.1);">
              <div style="font-size: 0.875rem; font-weight: 500; margin-bottom: 8px;">{{ t('leagues.whatsNext') }}</div>
              <n-list>
                <n-list-item>
                  <template #prefix>
                    <n-icon color="#2080f0"><TrophyIcon /></n-icon>
                  </template>
                  <div>{{ t('leagues.viewStandings') }}</div>
                </n-list-item>
                <n-list-item>
                  <template #prefix>
                    <n-icon color="#2080f0"><GameControllerIcon /></n-icon>
                  </template>
                  <div>{{ t('leagues.playGames') }}</div>
                </n-list-item>
                <n-list-item>
                  <template #prefix>
                    <n-icon color="#2080f0"><PeopleIcon /></n-icon>
                  </template>
                  <div>{{ t('leagues.inviteOthers') }}</div>
                </n-list-item>
              </n-list>
            </n-card>

            <n-space justify="center">
              <n-button type="primary" @click="goToLeague">
                <template #icon>
                  <n-icon><ArrowForwardIcon /></n-icon>
                </template>
                {{ t('leagues.goToLeague') }}
              </n-button>
              <n-button @click="goToHome">
                {{ t('leagues.goToHome') }}
              </n-button>
            </n-space>
          </div>

          <!-- Initial State (no token) -->
          <div v-else style="text-align: center; padding: 64px 0;">
            <n-icon size="80" color="#f0a020" style="margin-bottom: 16px; display: block;">
              <HelpCircleIcon />
            </n-icon>
            <div style="font-size: 1.125rem; font-weight: 500; margin-bottom: 8px;">{{ t('leagues.invalidInvitation') }}</div>
            <div style="margin-bottom: 16px;">
              {{ t('leagues.noToken') }}
            </div>

            <n-button type="primary" @click="goToHome">
              {{ t('leagues.goToHome') }}
            </n-button>
          </div>
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue';
import { NGrid, NGi, NCard, NIcon, NSpin, NButton, NSpace, NAlert, NDivider, NList, NListItem } from 'naive-ui';
import { 
  MailOpen as MailOpenIcon,
  LogIn as LogInIcon,
  CheckmarkCircle as PersonCheckIcon,
  AlertCircle as AlertCircleIcon,
  CheckmarkCircle as CheckCircleIcon,
  HelpCircle as HelpCircleIcon,
  ArrowForward as ArrowForwardIcon,
  Trophy as TrophyIcon,
  GameController as GameControllerIcon,
  People as PeopleIcon,
  LogoGoogle as LogoGoogleIcon,
  LogoDiscord as LogoDiscordIcon
} from '@vicons/ionicons5';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useLeagueStore } from '@/store/league';
import { useUserStore } from '@/store/user';
import LeagueApi from '@/api/LeagueApi';
import type { League, InvitationPreview } from '@/api/LeagueApi';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const leagueStore = useLeagueStore();
const userStore = useUserStore();

const loading = ref(false);
const loadingPreview = ref(false);
const success = ref(false);
const needsLogin = ref(false);
const alreadyMember = ref(false);
const alreadyMemberLeagueCode = ref<string | null>(null);
const error = ref<string | null>(null);
const league = ref<League | null>(null);
const preview = ref<InvitationPreview | null>(null);

const loadingMessage = computed(() => {
  if (loadingPreview.value) {
    return t('leagues.loadingInvitation');
  }
  return t('leagues.acceptingInvitation');
});

const loadPreview = async (token: string) => {
  try {
    loadingPreview.value = true;
    loading.value = true;
    preview.value = await LeagueApi.previewInvitation(token);
    
    // Check if invitation is still valid
    if (preview.value.status === 'expired') {
      error.value = t('leagues.invitationExpired');
    } else if (preview.value.status === 'used') {
      error.value = t('leagues.invitationUsed');
    }
  } catch (err) {
    console.error('Error loading invitation preview:', err);
    error.value = t('leagues.invitationNotFound');
  } finally {
    loadingPreview.value = false;
    loading.value = false;
  }
};

const acceptInvitation = async (token: string) => {
  loading.value = true;
  error.value = null;

  try {
    const result = await leagueStore.acceptInvitation(token);
    league.value = result.league;
    success.value = true;
  } catch (err) {
    if (err instanceof Error) {
      // Check for already member error with league code
      const errWithCode = err as Error & { leagueCode?: string };
      if (errWithCode.leagueCode) {
        alreadyMember.value = true;
        alreadyMemberLeagueCode.value = errWithCode.leagueCode;
      } else if (err.message.includes('already a member')) {
        alreadyMember.value = true;
      } else if (err.message.includes('404') || err.message.includes('not found')) {
        error.value = t('leagues.invitationNotFound');
      } else if (err.message.includes('expired')) {
        error.value = t('leagues.invitationExpired');
      } else if (err.message.includes('own invitation')) {
        error.value = t('leagues.cannotAcceptOwnInvitation');
      } else if (err.message.includes('already been used')) {
        error.value = t('leagues.invitationUsed');
      } else {
        error.value = err.message;
      }
    } else {
      error.value = t('leagues.error');
    }
    console.error('Error accepting invitation:', err);
  } finally {
    loading.value = false;
  }
};

const goToLeague = () => {
  if (league.value) {
    router.push({ name: 'LeagueDetails', params: { code: league.value.code } });
  }
};

const goToLeagueByCode = (code: string) => {
  router.push({ name: 'LeagueDetails', params: { code } });
};

const goToHome = () => {
  router.push({ name: 'Home' });
};

onMounted(async () => {
  const token = route.params.token as string;
  if (!token) {
    return;
  }

  // Load preview first (public, no auth required)
  await loadPreview(token);
  
  // If there was an error loading preview, stop here
  if (error.value) {
    return;
  }

  // Check if user is authenticated
  if (!userStore.isAuthenticated) {
    // Store the invitation URL in session storage so we can redirect back after login
    sessionStorage.setItem('invitation_return_url', route.fullPath);
    needsLogin.value = true;
    return;
  }

  await acceptInvitation(token);
});
</script>
