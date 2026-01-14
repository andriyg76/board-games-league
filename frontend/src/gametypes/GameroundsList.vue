<template>
  <n-grid :cols="24" :x-gap="16">
    <n-gi :span="24">
      <h2 style="font-size: 2rem; margin-bottom: 16px;">{{ $t('gameRounds.title') }}</h2>

      <n-alert v-if="error" type="error" style="margin-bottom: 16px;" closable @close="error = null">
        {{ error }}
      </n-alert>

      <n-spin v-if="loading" size="large" style="display: flex; justify-content: center; padding: 64px;" />

      <template v-else>
        <!-- Active Games Section -->
        <div v-if="activeRounds.length > 0" style="margin-bottom: 24px;">
          <h3 style="font-size: 1.25rem; font-weight: 500; margin-bottom: 8px;">{{ $t('common.inProgress') }}</h3>
          <n-list>
            <n-list-item v-for="round in activeRounds" :key="round.code">
              <div style="flex: 1;">
                <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 4px;">
                  <span style="font-weight: 500;">{{ round.name || $t('common.unknown') }}</span>
                  <n-tag :type="getStatusTagType(round.status)" size="small">
                    {{ getStatusLabel(round.status) }}
                  </n-tag>
                </div>
                <div style="font-size: 0.875rem; opacity: 0.7;">
                  {{ $t('gameRounds.started') }}: {{ formatDate(round.start_time) }}
                </div>
              </div>
              <template #suffix>
                <n-button type="primary" @click="continueRound(round)">
                  <template #icon>
                    <n-icon><PlayIcon /></n-icon>
                  </template>
                  {{ $t('gameRounds.continue') }}
                </n-button>
              </template>
            </n-list-item>
          </n-list>
        </div>

        <!-- Completed Games Section -->
        <div v-if="completedRounds.length > 0" style="margin-bottom: 24px;">
          <h3 style="font-size: 1.25rem; font-weight: 500; margin-bottom: 8px;">{{ $t('common.completed') }}</h3>
          <n-list>
            <n-list-item v-for="round in completedRounds" :key="round.code">
              <div style="flex: 1;">
                <div style="font-weight: 500; margin-bottom: 4px;">{{ round.name || $t('common.unknown') }}</div>
                <div style="font-size: 0.875rem; opacity: 0.7;">
                  {{ $t('gameRounds.started') }}: {{ formatDate(round.start_time) }}
                  {{ round.end_time ? ` | ${$t('gameRounds.ended')}: ${formatDate(round.end_time)}` : '' }}
                </div>
              </div>
              <template #suffix>
                <n-button quaternary @click="editRound(round)">
                  <template #icon>
                    <n-icon><EyeIcon /></n-icon>
                  </template>
                  {{ $t('gameRounds.view') }}
                </n-button>
              </template>
            </n-list-item>
          </n-list>
        </div>

        <!-- Empty State -->
        <n-alert v-if="gameRounds.length === 0" type="info" style="margin-bottom: 16px;">
          {{ $t('home.noGameRoundsYet') }}
        </n-alert>
      </template>

      <n-button type="primary" @click="createNewRound" style="margin-top: 16px;">
        <template #icon>
          <n-icon><AddIcon /></n-icon>
        </template>
        {{ $t('home.newGameRound') }}
      </n-button>
    </n-gi>
  </n-grid>

  <FinalizeGameDialog
    v-model="showFinalizeDialog"
    :round-code="selectedRoundCode"
    @finalized="handleFinalized"
  />
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
import { NGrid, NGi, NAlert, NSpin, NList, NListItem, NTag, NButton, NIcon } from 'naive-ui';
import { Play as PlayIcon, Eye as EyeIcon, Add as AddIcon } from '@vicons/ionicons5';
import { GameRoundView, GameRoundStatus } from './types';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import GameApi from '@/api/GameApi';
import FinalizeGameDialog from './FinalizeGameDialog.vue';

const router = useRouter();
const { t } = useI18n();

const gameRounds = ref<GameRoundView[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

const showFinalizeDialog = ref(false);
const selectedRoundCode = ref('');

// Separate active and completed rounds
const activeRounds = computed(() => 
  gameRounds.value.filter(r => r.status && r.status !== 'completed')
);

const completedRounds = computed(() => 
  gameRounds.value.filter(r => !r.status || r.status === 'completed')
);

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString();
};

const getStatusTagType = (status?: GameRoundStatus): 'default' | 'info' | 'success' | 'warning' | 'error' => {
  switch (status) {
    case 'players_selected': return 'info';
    case 'in_progress': return 'warning';
    case 'scoring': return 'success';
    case 'completed': return 'default';
    default: return 'default';
  }
};

const getStatusLabel = (status?: GameRoundStatus): string => {
  switch (status) {
    case 'players_selected': return t('gameRounds.statusPlayersSelected');
    case 'in_progress': return t('gameRounds.statusInProgress');
    case 'scoring': return t('gameRounds.statusScoring');
    case 'completed': return t('common.completed');
    default: return t('common.unknown');
  }
};

const loadGameRounds = async () => {
  loading.value = true;
  error.value = null;
  try {
    gameRounds.value = await GameApi.listGameRounds() as any;
  } catch (err) {
    console.error('Error fetching game rounds:', err);
    error.value = 'Failed to load game rounds';
  } finally {
    loading.value = false;
  }
};

const handleFinalized = async () => {
  await loadGameRounds();
};

const continueRound = (round: GameRoundView) => {
  router.push({ name: 'EditGameRound', params: { id: round.code }});
};

const editRound = (round: GameRoundView) => {
  // Use edit page for completed rounds
  router.push({ name: 'EditCompletedGameRound', params: { id: round.code }});
};

const createNewRound = () => {
  router.push({ name: 'NewGameRound' });
};

onMounted(async () => {
  await loadGameRounds();
});
</script>