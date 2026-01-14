<template>
  <n-modal v-model:show="dialog" preset="card" title="Finalize Game Round" style="max-width: 800px;" :mask-closable="false">
    <n-alert v-if="error" type="error" style="margin-bottom: 16px;" closable @close="error = null">
      {{ error }}
    </n-alert>

    <n-spin v-if="loading" size="large" style="display: flex; justify-content: center; padding: 64px;" />

    <div v-else-if="roundData">
      <p style="font-size: 1.125rem; font-weight: 500; margin-bottom: 16px;">{{ roundData.name }}</p>

      <n-card style="margin-bottom: 16px;">
        <template #header>
          Player Scores
        </template>
        <n-list>
          <n-list-item v-for="player in players" :key="player.user_id">
            <n-grid :cols="24" :x-gap="8">
              <n-gi :span="24" :responsive="{ m: 12 }">
                <div>
                  <div style="font-weight: 500;">{{ getPlayerName(player.user_id) }}</div>
                  <div style="font-size: 0.875rem; opacity: 0.7;">
                    <span v-if="player.team_name">Team: {{ player.team_name }}</span>
                    <span v-if="player.is_moderator">Moderator</span>
                  </div>
                </div>
              </n-gi>
              <n-gi :span="24" :responsive="{ m: 12 }">
                <n-input-number
                  v-if="!player.is_moderator"
                  v-model:value="playerScores[player.user_id]"
                  placeholder="Score"
                  :min="0"
                  size="small"
                  style="width: 100%;"
                />
                <span v-else style="color: #999;">N/A</span>
              </n-gi>
            </n-grid>
          </n-list-item>
        </n-list>
      </n-card>

      <n-card v-if="teams.length > 0">
        <template #header>
          Team Scores
        </template>
        <n-list>
          <n-list-item v-for="team in teams" :key="team">
            <n-grid :cols="24" :x-gap="8">
              <n-gi :span="24" :responsive="{ m: 12 }">
                <div style="font-weight: 500;">{{ team }}</div>
              </n-gi>
              <n-gi :span="24" :responsive="{ m: 12 }">
                <n-input-number
                  v-model:value="teamScores[team]"
                  placeholder="Team Score"
                  :min="0"
                  size="small"
                  style="width: 100%;"
                />
              </n-gi>
            </n-grid>
          </n-list-item>
        </n-list>
      </n-card>
    </div>

    <template #action>
      <div style="display: flex; justify-content: flex-end; gap: 8px;">
        <n-button quaternary @click="closeDialog" :disabled="submitting">
          Cancel
        </n-button>
        <n-button
          type="primary"
          @click="submitFinalization"
          :loading="submitting"
          :disabled="!canSubmit"
        >
          Finalize
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
import { NModal, NAlert, NSpin, NCard, NList, NListItem, NGrid, NGi, NInputNumber, NButton } from 'naive-ui';
import GameApi, { FinalizeGameRoundRequest } from '@/api/GameApi';
import { GameRoundView } from './types';

const props = defineProps<{
  modelValue: boolean;
  roundCode: string;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'finalized'): void;
}>();

const dialog = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
});

const roundData = ref<GameRoundView | null>(null);
const loading = ref(false);
const submitting = ref(false);
const error = ref<string | null>(null);

const playerScores = ref<Record<string, number>>({});
const teamScores = ref<Record<string, number>>({});
const playerNames = ref<Record<string, string>>({});

const players = computed(() => roundData.value?.players || []);

const teams = computed(() => {
  if (!roundData.value) return [];
  const uniqueTeams = new Set(
    roundData.value.players
      .filter(p => p.team_name)
      .map(p => p.team_name!)
  );
  return Array.from(uniqueTeams);
});

const canSubmit = computed(() => {
  return Object.keys(playerScores.value).length > 0;
});

const getPlayerName = (userId: string) => {
  return playerNames.value[userId] || userId;
};

const loadGameRound = async () => {
  if (!props.roundCode) return;

  loading.value = true;
  error.value = null;

  try {
    roundData.value = await GameApi.getGameRound(props.roundCode) as any;

    // Initialize player scores
    playerScores.value = {};
    roundData.value?.players.forEach(player => {
      if (!player.is_moderator) {
        playerScores.value[player.user_id] = player.score || 0;
      }
    });

    // Initialize team scores
    teamScores.value = {};
    teams.value.forEach(team => {
      teamScores.value[team] = 0;
    });

    // Load player names
    await loadPlayerNames();
  } catch (err) {
    console.error('Error loading game round:', err);
    error.value = 'Failed to load game round data';
  } finally {
    loading.value = false;
  }
};

const loadPlayerNames = async () => {
  try {
    const players = await GameApi.listPlayers();
    playerNames.value = {};
    players.forEach(player => {
      playerNames.value[player.code] = player.alias;
    });
  } catch (err) {
    console.error('Error loading player names:', err);
  }
};

const submitFinalization = async () => {
  if (!props.roundCode) return;

  submitting.value = true;
  error.value = null;

  try {
    const finalizationData: FinalizeGameRoundRequest = {
      player_scores: playerScores.value,
    };

    if (teams.value.length > 0) {
      finalizationData.team_scores = teamScores.value;
    }

    await GameApi.finalizeGameRound(props.roundCode, finalizationData);

    emit('finalized');
    closeDialog();
  } catch (err) {
    console.error('Error finalizing game round:', err);
    error.value = 'Failed to finalize game round';
  } finally {
    submitting.value = false;
  }
};

const closeDialog = () => {
  dialog.value = false;
  roundData.value = null;
  playerScores.value = {};
  teamScores.value = {};
  error.value = null;
};

watch(() => props.modelValue, (newValue) => {
  if (newValue) {
    loadGameRound();
  }
});
</script>
