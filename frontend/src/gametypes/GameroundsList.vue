<template>
  <v-container>
    <v-row>
      <v-col>
        <h2>{{ $t('gameRounds.title') }}</h2>

        <v-alert v-if="error" type="error" class="mb-4">
          {{ error }}
        </v-alert>

        <v-progress-circular v-if="loading" indeterminate color="primary" />

        <template v-else>
          <!-- Active Games Section -->
          <div v-if="activeRounds.length > 0" class="mb-6">
            <h3 class="text-h6 mb-2">{{ $t('common.inProgress') }}</h3>
            <v-list>
              <v-list-item v-for="round in activeRounds" :key="round.code">
                <template v-slot:default>
                  <v-list-item-title>
                    {{ round.name || $t('common.unknown') }}
                    <v-chip :color="getStatusColor(round.status)" size="small" class="ml-2">
                      {{ getStatusLabel(round.status) }}
                    </v-chip>
                  </v-list-item-title>
                  <v-list-item-subtitle>
                    {{ $t('gameRounds.started') }}: {{ formatDate(round.start_time) }}
                  </v-list-item-subtitle>
                </template>
                <template v-slot:append>
                  <v-btn @click="continueRound(round)" color="primary" variant="elevated">
                    <v-icon start>mdi-play</v-icon>
                    {{ $t('gameRounds.continue') }}
                  </v-btn>
                </template>
              </v-list-item>
            </v-list>
          </div>

          <!-- Completed Games Section -->
          <div v-if="completedRounds.length > 0">
            <h3 class="text-h6 mb-2">{{ $t('common.completed') }}</h3>
            <v-list>
              <v-list-item v-for="round in completedRounds" :key="round.code">
                <template v-slot:default>
                  <v-list-item-title>{{ round.name || $t('common.unknown') }}</v-list-item-title>
                  <v-list-item-subtitle>
                    {{ $t('gameRounds.started') }}: {{ formatDate(round.start_time) }}
                    {{ round.end_time ? ` | ${$t('gameRounds.ended')}: ${formatDate(round.end_time)}` : '' }}
                  </v-list-item-subtitle>
                </template>
                <template v-slot:append>
                  <v-btn @click="editRound(round)" color="grey" variant="text">
                    <v-icon start>mdi-eye</v-icon>
                    {{ $t('gameRounds.view') }}
                  </v-btn>
                </template>
              </v-list-item>
            </v-list>
          </div>

          <!-- Empty State -->
          <v-alert v-if="gameRounds.length === 0" type="info">
            {{ $t('home.noGameRoundsYet') }}
          </v-alert>
        </template>

        <v-btn @click="createNewRound" color="primary" class="mt-4">
          <v-icon start>mdi-plus</v-icon>
          {{ $t('home.newGameRound') }}
        </v-btn>
      </v-col>
    </v-row>

    <FinalizeGameDialog
      v-model="showFinalizeDialog"
      :round-code="selectedRoundCode"
      @finalized="handleFinalized"
    />
  </v-container>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
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

const getStatusColor = (status?: GameRoundStatus): string => {
  switch (status) {
    case 'players_selected': return 'blue';
    case 'in_progress': return 'orange';
    case 'scoring': return 'green';
    case 'completed': return 'grey';
    default: return 'grey';
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
    gameRounds.value = await GameApi.listGameRounds();
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
  router.push({ name: 'EditGameRound', params: { id: round.code }});
};

const createNewRound = () => {
  router.push({ name: 'NewGameRound' });
};

onMounted(async () => {
  await loadGameRounds();
});
</script>