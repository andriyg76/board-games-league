<template>
  <v-container>
    <v-row>
      <v-col>
        <h2>{{ $t('gameRounds.title') }}</h2>

        <div v-if="loading" class="text-center pa-4">
          <v-progress-circular indeterminate color="primary" size="64" />
          <p class="mt-4">{{ $t('common.loading') }}</p>
        </div>

        <v-form v-else @submit.prevent="saveRound">
          <v-text-field
            v-model="round.name"
            :label="$t('game.roundName')"
            variant="outlined"
            class="mb-4"
          />

          <v-list v-if="round.players.length > 0">
            <v-subheader>{{ $t('game.selectedPlayers') }}</v-subheader>
            <v-list-item v-for="(player, index) in round.players" :key="index">
              <v-row align="center">
                <v-col cols="4">
                  <span>{{ getPlayerAlias(player) }}</span>
                </v-col>
                <v-col cols="3">
                  <v-text-field
                    v-model.number="player.score"
                    :label="$t('leagues.points')"
                    type="number"
                    density="compact"
                    variant="outlined"
                    hide-details
                  />
                </v-col>
                <v-col cols="2">
                  <v-text-field
                    v-model.number="player.position"
                    :label="$t('game.position')"
                    type="number"
                    density="compact"
                    variant="outlined"
                    hide-details
                  />
                </v-col>
                <v-col cols="3">
                  <v-checkbox
                    v-model="player.is_moderator"
                    :label="$t('roleTypes.moderator')"
                    hide-details
                    density="compact"
                  />
                </v-col>
              </v-row>
            </v-list-item>
          </v-list>

          <div class="d-flex gap-2 mt-4">
            <v-btn variant="text" @click="goBack">
              {{ $t('common.cancel') }}
            </v-btn>
            <v-spacer />
            <v-btn
              type="submit"
              color="success"
              :loading="saving"
            >
              <v-icon start>mdi-content-save</v-icon>
              {{ $t('common.save') }}
            </v-btn>
          </div>
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useGameStore } from '@/store/game';
import { usePlayerStore } from '@/store/player';
import GameApi, { GameRound, GameRoundPlayer } from '@/api/GameApi';

const props = defineProps<{
  id: string;
}>();

const router = useRouter();
const gameStore = useGameStore();
const playerStore = usePlayerStore();

const loading = ref(false);
const saving = ref(false);
const round = ref<GameRound>({
  code: '',
  name: '',
  game_type: '',
  start_time: new Date().toISOString(),
  players: [],
  version: 0
});

const getPlayerAlias = (player: GameRoundPlayer): string => {
  // Try to find player in player store
  const found = playerStore.players?.find(p => p.code === player.user_id);
  if (found) return found.alias;
  
  // Fallback to membership_id
  return player.membership_id || player.user_id || 'Unknown';
};

const loadRound = async () => {
  loading.value = true;
  try {
    const loadedRound = await GameApi.getGameRound(props.id);
    round.value = {
      code: props.id,
      name: loadedRound.name,
      game_type: loadedRound.game_type,
      start_time: loadedRound.start_time,
      players: loadedRound.players,
      version: loadedRound.version
    };
  } catch (error) {
    console.error('Error loading game round:', error);
  } finally {
    loading.value = false;
  }
};

const saveRound = async () => {
  saving.value = true;
  try {
    await gameStore.updateRound(round.value);
    await router.push({ name: 'GameRounds' });
  } catch (error) {
    console.error('Error saving game round:', error);
  } finally {
    saving.value = false;
  }
};

const goBack = () => {
  router.push({ name: 'GameRounds' });
};

onMounted(async () => {
  // Load players for alias lookup
  if (!playerStore.players || playerStore.players.length === 0) {
    try {
      const players = await GameApi.listPlayers();
      playerStore.players = players;
    } catch (error) {
      console.error('Failed to load players:', error);
    }
  }

  await loadRound();
});
</script>

<style scoped>
.gap-2 {
  gap: 0.5rem;
}
</style>
