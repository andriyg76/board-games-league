<template>
  <div>
    <n-grid :cols="24" :x-gap="16" style="margin-bottom: 24px;">
      <n-gi :span="24">
        <h1 style="font-size: 2rem; margin-bottom: 16px;">{{ t('home.title') }}</h1>
        <p style="font-size: 1.25rem; color: rgba(0, 0, 0, 0.6); margin-bottom: 24px;">
          {{ t('home.welcome') }}
        </p>
      </n-gi>
    </n-grid>

    <n-spin v-if="loading" size="large" style="display: flex; justify-content: center; padding: 64px;" />

    <n-grid v-else :cols="24" :x-gap="16" style="margin-bottom: 24px;">
      <n-gi :span="24" :responsive="{ m: 8 }">
        <n-card>
          <div style="font-size: 0.75rem; text-transform: uppercase; margin-bottom: 4px; opacity: 0.7;">{{ t('home.totalGameRounds') }}</div>
          <div style="font-size: 2rem; font-weight: 500;">{{ totalRounds }}</div>
          <div style="font-size: 0.75rem; opacity: 0.7;">{{ t('common.allTime') }}</div>
        </n-card>
      </n-gi>

      <n-gi :span="24" :responsive="{ m: 8 }">
        <n-card>
          <div style="font-size: 0.75rem; text-transform: uppercase; margin-bottom: 4px; opacity: 0.7;">{{ t('home.activeGames') }}</div>
          <div style="font-size: 2rem; font-weight: 500;">{{ activeRounds }}</div>
          <div style="font-size: 0.75rem; opacity: 0.7;">{{ t('common.inProgress') }}</div>
        </n-card>
      </n-gi>

      <n-gi :span="24" :responsive="{ m: 8 }">
        <n-card>
          <div style="font-size: 0.75rem; text-transform: uppercase; margin-bottom: 4px; opacity: 0.7;">{{ t('home.gameTypes') }}</div>
          <div style="font-size: 2rem; font-weight: 500;">{{ totalGameTypes }}</div>
          <div style="font-size: 0.75rem; opacity: 0.7;">{{ t('common.available') }}</div>
        </n-card>
      </n-gi>
    </n-grid>

    <n-grid :cols="24" :x-gap="16" style="margin-bottom: 24px;">
      <n-gi :span="24">
        <n-card>
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span style="font-size: 1.25rem; font-weight: 500;">{{ t('home.recentGameRounds') }}</span>
              <n-button quaternary @click="navigateToAllRounds">
                {{ t('common.viewAll') }}
              </n-button>
            </div>
          </template>
          <n-divider />

          <n-alert v-if="recentRounds.length === 0" type="info" style="margin-top: 16px;">
            {{ t('home.noGameRoundsYet') }}
          </n-alert>

          <n-list v-else>
            <n-list-item
              v-for="round in recentRounds"
              :key="round.code"
              clickable
              @click="navigateToRound(round.code)"
            >
              <template #prefix>
                <n-icon v-if="round.end_time" color="#18a058" size="20">
                  <CheckCircleIcon />
                </n-icon>
                <n-icon v-else color="#2080f0" size="20">
                  <PlayCircleIcon />
                </n-icon>
              </template>
              <div>
                <div style="font-weight: 500;">{{ round.name }}</div>
                <div style="font-size: 0.875rem; opacity: 0.7;">
                  {{ formatDate(round.start_time) }}
                  <span v-if="round.end_time"> - {{ t('common.completed') }}</span>
                  <span v-else> - {{ t('common.inProgress') }}</span>
                </div>
              </div>
            </n-list-item>
          </n-list>
        </n-card>
      </n-gi>
    </n-grid>

    <n-grid :cols="24" :x-gap="16">
      <n-gi :span="24" :responsive="{ m: 12 }">
        <n-card style="text-align: center; padding: 24px;" theme-overrides="{ colorPrimary: '#2080f0' }">
          <div style="font-size: 1.125rem; font-weight: 500; margin-bottom: 8px;">{{ t('home.createNewGameRound') }}</div>
          <div style="margin-bottom: 16px; opacity: 0.8;">{{ t('home.startTracking') }}</div>
          <n-button type="primary" size="large" @click="navigateToNewRound">
            <template #icon>
              <n-icon><AddCircleIcon /></n-icon>
            </template>
            {{ t('home.newGameRound') }}
          </n-button>
        </n-card>
      </n-gi>

      <n-gi :span="24" :responsive="{ m: 12 }">
        <n-card style="text-align: center; padding: 24px;" theme-overrides="{ colorPrimary: '#18a058' }">
          <div style="font-size: 1.125rem; font-weight: 500; margin-bottom: 8px;">{{ t('home.manageGameTypes') }}</div>
          <div style="margin-bottom: 16px; opacity: 0.8;">{{ t('home.configureCollection') }}</div>
          <n-button type="success" size="large" @click="navigateToGameTypes">
            <template #icon>
              <n-icon><DiceIcon /></n-icon>
            </template>
            {{ t('gameTypes.title') }}
          </n-button>
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue';
import { NGrid, NGi, NCard, NSpin, NButton, NList, NListItem, NIcon, NAlert, NDivider } from 'naive-ui';
import { CheckmarkCircle as CheckCircleIcon, PlayCircleOutline as PlayCircleIcon, AddCircleOutline as AddCircleIcon } from '@vicons/ionicons5';
import { Dice as DiceIcon } from '@vicons/ionicons5';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import GameApi from '@/api/GameApi';
import { GameRoundView } from '@/gametypes/types';
import { GameType } from '@/api/GameApi';

const { t, locale } = useI18n();
const router = useRouter();

const gameRounds = ref<GameRoundView[]>([]);
const gameTypes = ref<GameType[]>([]);
const loading = ref(true);

const recentRounds = computed(() => gameRounds.value.slice(0, 5));

const totalRounds = computed(() => gameRounds.value.length);

const activeRounds = computed(() =>
  gameRounds.value.filter(round => !round.end_time).length
);

const totalGameTypes = computed(() => gameTypes.value.length);

const formatDate = (dateStr: string) => {
  const localeMap: Record<string, string> = { 'uk': 'uk-UA', 'en': 'en-US', 'et': 'et-EE' };
  return new Date(dateStr).toLocaleDateString(localeMap[locale.value] || 'en-US');
};

const navigateToAllRounds = () => {
  router.push({ name: 'GameRounds' });
};

const navigateToRound = (code: string) => {
  router.push({ name: 'EditGameRound', params: { id: code } });
};

const navigateToNewRound = () => {
  router.push({ name: 'NewGameRound' });
};

const navigateToGameTypes = () => {
  router.push('/ui/admin/game-types');
};

onMounted(async () => {
  loading.value = true;
  try {
    const [roundsData, typesData] = await Promise.all([
      GameApi.listGameRounds(),
      GameApi.getGameTypes()
    ]);
    gameRounds.value = roundsData as GameRoundView[];
    gameTypes.value = typesData;
  } catch (error) {
    console.error('Error loading dashboard data:', error);
  } finally {
    loading.value = false;
  }
});
</script>
