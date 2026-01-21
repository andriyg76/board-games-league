<template>
  <div class="mobile-league-home">
    <n-spin v-if="loading" size="large" />

    <n-alert v-else-if="errorMessage" type="error">
      {{ errorMessage }}
    </n-alert>

    <div v-else>
      <div class="mobile-league-home__header">
        <h1 class="mobile-league-home__title">{{ league?.name || 'League' }}</h1>
      </div>

      <n-button type="primary" size="large" block @click="startNewGame">
        {{ t('gameRounds.start') }}
      </n-button>

      <div class="mobile-league-home__section">
        <h2 class="mobile-league-home__section-title">{{ t('gameRounds.list') }}</h2>
        <n-alert v-if="activeRounds.length === 0" type="info">
          {{ t('home.noGameRoundsYet') }}
        </n-alert>
        <n-list v-else>
          <n-list-item
            v-for="round in activeRounds"
            :key="round.code"
            clickable
            @click="continueGame(round.code)"
          >
            <div class="mobile-league-home__round">
              <div class="mobile-league-home__round-name">{{ round.name }}</div>
              <div class="mobile-league-home__round-date">{{ formatDate(round.start_time) }}</div>
            </div>
          </n-list-item>
        </n-list>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { NSpin, NAlert, NButton, NList, NListItem } from 'naive-ui';
import GameApi, { GameRound } from '@/api/GameApi';
import { useLeagueStore } from '@/store/league';
import { useUserStore } from '@/store/user';

const { t, locale } = useI18n();
const router = useRouter();
const leagueStore = useLeagueStore();
const userStore = useUserStore();

const loading = ref(true);
const errorMessage = ref<string | null>(null);
const rounds = ref<GameRound[]>([]);

const league = computed(() => leagueStore.currentLeague);
const activeRounds = computed(() => rounds.value.filter((round) => !round.end_time));

const formatDate = (dateStr: string) => {
  const localeMap: Record<string, string> = { uk: 'uk-UA', en: 'en-US', et: 'et-EE' };
  return new Date(dateStr).toLocaleDateString(localeMap[locale.value] || 'en-US');
};

const startNewGame = () => {
  router.push({ name: 'MobileGameStart' });
};

const continueGame = (code?: string) => {
  if (!code) return;
  router.push({ name: 'MobileGameFlow', params: { code } });
};

const loadLeagueHome = async () => {
  if (!userStore.loggedIn) {
    await router.replace({ name: 'MobileLogin', query: { redirect: '/m/league' } });
    return;
  }

  const leagueCode = leagueStore.currentLeagueCode;
  if (!leagueCode) {
    await router.replace({ name: 'MobileEntry' });
    return;
  }

  if (!league.value || league.value.code !== leagueCode) {
    await leagueStore.setCurrentLeague(leagueCode);
  }

  rounds.value = await GameApi.listLeagueGameRounds(leagueCode);
};

onMounted(async () => {
  loading.value = true;
  errorMessage.value = null;
  try {
    await loadLeagueHome();
  } catch (error) {
    console.error('Failed to load league home:', error);
    errorMessage.value = 'Unable to load league data.';
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.mobile-league-home {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 24px;
}

.mobile-league-home__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.mobile-league-home__title {
  font-size: 1.5rem;
  margin: 0;
}

.mobile-league-home__section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mobile-league-home__section-title {
  font-size: 1.1rem;
  margin: 0;
}

.mobile-league-home__round {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mobile-league-home__round-name {
  font-weight: 500;
}

.mobile-league-home__round-date {
  font-size: 0.85rem;
  opacity: 0.7;
}
</style>
