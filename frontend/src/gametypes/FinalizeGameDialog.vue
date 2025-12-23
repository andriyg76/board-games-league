<template>
  <v-dialog v-model="dialog" max-width="800" persistent>
    <v-card>
      <v-card-title class="text-h5">
        Finalize Game Round
      </v-card-title>

      <v-card-text>
        <v-alert v-if="error" type="error" class="mb-4">
          {{ error }}
        </v-alert>

        <div v-if="loading" class="text-center pa-4">
          <v-progress-circular indeterminate color="primary" />
        </div>

        <div v-else-if="roundData">
          <p class="text-subtitle-1 mb-4">{{ roundData.name }}</p>

          <v-list>
            <v-subheader>Player Scores</v-subheader>
            <v-list-item v-for="player in players" :key="player.user_id">
              <v-row align="center">
                <v-col cols="6">
                  <v-list-item-title>{{ getPlayerName(player.user_id) }}</v-list-item-title>
                  <v-list-item-subtitle v-if="player.team_name">
                    Team: {{ player.team_name }}
                  </v-list-item-subtitle>
                  <v-list-item-subtitle v-if="player.is_moderator">
                    Moderator
                  </v-list-item-subtitle>
                </v-col>
                <v-col cols="6">
                  <v-text-field
                    v-if="!player.is_moderator"
                    v-model.number="playerScores[player.user_id]"
                    label="Score"
                    type="number"
                    density="compact"
                    hide-details
                  />
                  <span v-else class="text-disabled">N/A</span>
                </v-col>
              </v-row>
            </v-list-item>
          </v-list>

          <v-list v-if="teams.length > 0" class="mt-4">
            <v-subheader>Team Scores</v-subheader>
            <v-list-item v-for="team in teams" :key="team">
              <v-row align="center">
                <v-col cols="6">
                  <v-list-item-title>{{ team }}</v-list-item-title>
                </v-col>
                <v-col cols="6">
                  <v-text-field
                    v-model.number="teamScores[team]"
                    label="Team Score"
                    type="number"
                    density="compact"
                    hide-details
                  />
                </v-col>
              </v-row>
            </v-list-item>
          </v-list>
        </div>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn
          color="grey"
          variant="text"
          @click="closeDialog"
          :disabled="submitting"
        >
          Cancel
        </v-btn>
        <v-btn
          color="primary"
          variant="elevated"
          @click="submitFinalization"
          :loading="submitting"
          :disabled="!canSubmit"
        >
          Finalize
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
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
    roundData.value = await GameApi.getGameRound(props.roundCode);

    // Initialize player scores
    playerScores.value = {};
    roundData.value.players.forEach(player => {
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
