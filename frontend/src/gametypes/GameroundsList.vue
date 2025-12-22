<template>
  <v-container>
    <v-row>
      <v-col>
        <h2>Game Rounds</h2>

        <v-alert v-if="error" type="error" class="mb-4">
          {{ error }}
        </v-alert>

        <v-progress-circular v-if="loading" indeterminate color="primary" />

        <v-list v-else-if="gameRounds.length > 0">
          <v-list-item v-for="round in gameRounds" :key="round.code">
            <template v-slot:default>
              <v-list-item-title>{{ round.name }}</v-list-item-title>
              <v-list-item-subtitle>
                Started: {{ formatDate(round.start_time) }}
                {{ round.end_time ? `| Ended: ${formatDate(round.end_time)}` : '' }}
              </v-list-item-subtitle>
            </template>
            <template v-slot:append>
              <v-btn @click="editRound(round)" color="primary" class="mr-2">Edit</v-btn>
              <v-btn v-if="!round.end_time" @click="finalizeRound(round.code)" color="success">Finalize</v-btn>
            </template>
          </v-list-item>
        </v-list>

        <v-alert v-else type="info">
          No game rounds found. Create one to get started!
        </v-alert>

        <v-btn @click="createNewRound" color="primary" class="mt-4">
          Create New Round
        </v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { GameRoundView } from './types';
import { useRouter } from 'vue-router';
import GameApi from '@/api/GameApi';

const gameRounds = ref<GameRoundView[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString();
};

const finalizeRound = async (code: string) => {
  try {
    // Navigate to the finalization page or implement inline finalization
    router.push({ name: 'EditGameRound', params: { id: code }});
  } catch (err) {
    console.error('Error finalizing round:', err);
    error.value = 'Failed to finalize game round';
  }
};

onMounted(async () => {
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
});

const router = useRouter();

const editRound = (round: GameRoundView) => {
  router.push({ name: 'EditGameRound', params: { id: round.code }});
};

const createNewRound = () => {
  router.push({ name: 'NewGameRound' });
};
</script>