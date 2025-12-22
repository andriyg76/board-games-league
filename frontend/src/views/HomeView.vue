<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h3 mb-4">Board Games League</h1>
        <p class="text-h6 text-medium-emphasis mb-6">
          Welcome to your board games tracking dashboard
        </p>
      </v-col>
    </v-row>

    <v-row v-if="loading">
      <v-col cols="12" class="text-center">
        <v-progress-circular indeterminate color="primary" size="64" />
      </v-col>
    </v-row>

    <v-row v-else>
      <v-col cols="12" md="4">
        <v-card elevation="2">
          <v-card-text>
            <div class="text-overline mb-1">Total Game Rounds</div>
            <div class="text-h4">{{ totalRounds }}</div>
            <div class="text-caption">All time</div>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="4">
        <v-card elevation="2">
          <v-card-text>
            <div class="text-overline mb-1">Active Games</div>
            <div class="text-h4">{{ activeRounds }}</div>
            <div class="text-caption">In progress</div>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="4">
        <v-card elevation="2">
          <v-card-text>
            <div class="text-overline mb-1">Game Types</div>
            <div class="text-h4">{{ totalGameTypes }}</div>
            <div class="text-caption">Available</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-4">
      <v-col cols="12">
        <v-card elevation="2">
          <v-card-title>
            <span class="text-h5">Recent Game Rounds</span>
            <v-spacer />
            <v-btn
              color="primary"
              @click="navigateToAllRounds"
              variant="text"
            >
              View All
            </v-btn>
          </v-card-title>
          <v-divider />

          <v-card-text v-if="recentRounds.length === 0">
            <v-alert type="info" variant="tonal">
              No game rounds yet. Start by creating your first game round!
            </v-alert>
          </v-card-text>

          <v-list v-else>
            <v-list-item
              v-for="round in recentRounds"
              :key="round.code"
              @click="navigateToRound(round.code)"
            >
              <template v-slot:prepend>
                <v-icon v-if="round.end_time" color="success">mdi-check-circle</v-icon>
                <v-icon v-else color="primary">mdi-play-circle</v-icon>
              </template>

              <v-list-item-title>{{ round.name }}</v-list-item-title>
              <v-list-item-subtitle>
                {{ formatDate(round.start_time) }}
                <span v-if="round.end_time"> - Completed</span>
                <span v-else> - In Progress</span>
              </v-list-item-subtitle>
            </v-list-item>
          </v-list>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-4">
      <v-col cols="12" md="6">
        <v-card elevation="2" class="text-center pa-4" color="primary" variant="tonal">
          <v-card-title class="text-h6">Create New Game Round</v-card-title>
          <v-card-text>
            Start tracking a new board game session
          </v-card-text>
          <v-btn
            color="primary"
            size="large"
            @click="navigateToNewRound"
          >
            <v-icon start>mdi-plus-circle</v-icon>
            New Game Round
          </v-btn>
        </v-card>
      </v-col>

      <v-col cols="12" md="6">
        <v-card elevation="2" class="text-center pa-4" color="secondary" variant="tonal">
          <v-card-title class="text-h6">Manage Game Types</v-card-title>
          <v-card-text>
            Configure your board game collection
          </v-card-text>
          <v-btn
            color="secondary"
            size="large"
            @click="navigateToGameTypes"
          >
            <v-icon start>mdi-dice-multiple</v-icon>
            Game Types
          </v-btn>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import GameApi from '@/api/GameApi';
import { GameRoundView } from '@/gametypes/types';
import { GameType } from '@/api/GameApi';

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
  return new Date(dateStr).toLocaleDateString();
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
    gameRounds.value = roundsData;
    gameTypes.value = typesData;
  } catch (error) {
    console.error('Error loading dashboard data:', error);
  } finally {
    loading.value = false;
  }
});
</script>
