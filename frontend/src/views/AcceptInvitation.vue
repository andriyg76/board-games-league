<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" md="6">
        <v-card elevation="4">
          <v-card-title class="text-h5 text-center bg-primary pa-6">
            <v-icon size="large" class="mr-2">mdi-email-open</v-icon>
            {{ t('leagues.invitation') }}
          </v-card-title>

          <!-- Loading State -->
          <v-card-text v-if="loading" class="text-center py-8">
            <v-progress-circular
              indeterminate
              color="primary"
              size="64"
              class="mb-4"
            />
            <div class="text-h6">{{ t('leagues.acceptingInvitation') }}</div>
          </v-card-text>

          <!-- Login Required State -->
          <v-card-text v-else-if="needsLogin" class="text-center py-8">
            <v-icon
              size="80"
              color="info"
              class="mb-4"
            >
              mdi-login
            </v-icon>
            <div class="text-h6 mb-2">{{ t('leagues.pleaseLoginFirst') }}</div>
            <div class="text-body-1 mb-4">
              {{ t('leagues.loginToAcceptInvitation') }}
            </div>

            <div class="d-flex gap-2 justify-center flex-wrap">
              <v-btn
                color="primary"
                variant="flat"
                href="/oauth/login/google"
              >
                <v-icon start>mdi-google</v-icon>
                {{ t('auth.loginWithGoogle') }}
              </v-btn>
              <v-btn
                color="indigo"
                variant="flat"
                href="/oauth/login/discord"
              >
                <v-icon start>mdi-discord</v-icon>
                {{ t('auth.loginWithDiscord') }}
              </v-btn>
            </div>
          </v-card-text>

          <!-- Error State -->
          <v-card-text v-else-if="error" class="text-center py-8">
            <v-icon
              size="80"
              color="error"
              class="mb-4"
            >
              mdi-alert-circle
            </v-icon>
            <div class="text-h6 mb-2">{{ t('leagues.error') }}</div>
            <v-alert type="error" variant="tonal">
              {{ error }}
            </v-alert>

            <v-btn
              color="primary"
              class="mt-4"
              @click="goToHome"
            >
              {{ t('leagues.goToHome') }}
            </v-btn>
          </v-card-text>

          <!-- Success State -->
          <v-card-text v-else-if="success && league" class="text-center py-8">
            <v-icon
              size="80"
              color="success"
              class="mb-4"
            >
              mdi-check-circle
            </v-icon>
            <div class="text-h6 mb-2">{{ t('leagues.congratulations') }}</div>
            <div class="text-body-1 mb-4">
              {{ t('leagues.joinedLeague') }} <strong>{{ league.name }}</strong>
            </div>

            <v-divider class="my-4" />

            <v-card variant="tonal" color="primary" class="mb-4">
              <v-card-text>
                <div class="text-subtitle-2 mb-2">{{ t('leagues.whatsNext') }}</div>
                <v-list density="compact" bg-color="transparent">
                  <v-list-item>
                    <template v-slot:prepend>
                      <v-icon>mdi-trophy</v-icon>
                    </template>
                    <v-list-item-title>{{ t('leagues.viewStandings') }}</v-list-item-title>
                  </v-list-item>
                  <v-list-item>
                    <template v-slot:prepend>
                      <v-icon>mdi-gamepad-variant</v-icon>
                    </template>
                    <v-list-item-title>{{ t('leagues.playGames') }}</v-list-item-title>
                  </v-list-item>
                  <v-list-item>
                    <template v-slot:prepend>
                      <v-icon>mdi-account-group</v-icon>
                    </template>
                    <v-list-item-title>{{ t('leagues.inviteOthers') }}</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-card-text>
            </v-card>

            <div class="d-flex gap-2 justify-center">
              <v-btn
                color="primary"
                variant="flat"
                @click="goToLeague"
              >
                <v-icon start>mdi-arrow-right</v-icon>
                {{ t('leagues.goToLeague') }}
              </v-btn>
              <v-btn
                variant="outlined"
                @click="goToHome"
              >
                {{ t('leagues.goToHome') }}
              </v-btn>
            </div>
          </v-card-text>

          <!-- Initial State (no token) -->
          <v-card-text v-else class="text-center py-8">
            <v-icon
              size="80"
              color="warning"
              class="mb-4"
            >
              mdi-help-circle
            </v-icon>
            <div class="text-h6 mb-2">{{ t('leagues.invalidInvitation') }}</div>
            <div class="text-body-1 mb-4">
              {{ t('leagues.noToken') }}
            </div>

            <v-btn
              color="primary"
              @click="goToHome"
            >
              {{ t('leagues.goToHome') }}
            </v-btn>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useLeagueStore } from '@/store/league';
import { useUserStore } from '@/store/user';
import type { League } from '@/api/LeagueApi';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const leagueStore = useLeagueStore();
const userStore = useUserStore();

const loading = ref(false);
const success = ref(false);
const needsLogin = ref(false);
const error = ref<string | null>(null);
const league = ref<League | null>(null);

const acceptInvitation = async (token: string) => {
  loading.value = true;
  error.value = null;

  try {
    const result = await leagueStore.acceptInvitation(token);
    league.value = result.league;
    success.value = true;
  } catch (err) {
    if (err instanceof Error) {
      if (err.message.includes('404')) {
        error.value = t('leagues.invitationNotFound');
      } else if (err.message.includes('expired')) {
        error.value = t('leagues.invitationExpired');
      } else if (err.message.includes('own invitation')) {
        error.value = t('leagues.cannotAcceptOwnInvitation');
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

const goToHome = () => {
  router.push({ name: 'Home' });
};

onMounted(async () => {
  const token = route.params.token as string;
  if (!token) {
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

<style scoped>
.gap-2 {
  gap: 8px;
}
</style>
