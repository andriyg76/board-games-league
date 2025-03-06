<template>
  <v-container>
    <v-row>
      <v-col>
        <h2>{{ isEditing ? 'Edit Game Round' : 'New Game Round' }}</h2>
        <v-form @submit.prevent="saveRound">
          <v-row>
            <v-col cols="9">
              <v-select
                  v-model="round.game_type"
                  :items="gameTypes"
                  item-title="name"
                  item-value="code"
                  label="Game Type"
                  required
                  :disabled="isEditing"
              />
            </v-col>
            <v-col cols="3" v-if="round.game_type && !isEditing">
              <v-btn
                  color="primary"
                  @click="startRound"
                  :loading="loading"
              >
                Start
              </v-btn>
            </v-col>
          </v-row>

          <v-text-field
              v-if="round.players.length > 0"
              v-model="round.name"
              label="Round Name"
              required
          />

          <v-list v-if="round.players.length > 0">
            <v-subheader>Players</v-subheader>
            <v-list-item v-for="(player, index) in round.players" :key="index">
              <v-row>
                <v-col cols="6">
                  <v-select
                      v-model="player.user_id"
                      :items="players"
                      item-title="alias"
                      item-value="code"
                      label="Player"
                      required
                  />
                </v-col>
                <v-col cols="3">
                  <v-checkbox
                      v-model="player.is_moderator"
                      label="Moderator"
                  />
                </v-col>
                <v-col cols="3" v-if="selectedGameType?.teams?.length">
                  <v-select
                      v-model="player.team_name"
                      :items="selectedGameType.teams"
                      item-title="name"
                      item-value="name"
                      label="Team"
                  />
                </v-col>
              </v-row>
            </v-list-item>
          </v-list>

          <v-btn
              v-if="round.players.length > 0"
              type="submit"
              color="success"
              class="mt-4"
          >
            Save
          </v-btn>
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
import { useGameStore } from '@/store/game';
import { usePlayerStore } from '@/store/player';
import { useRouter } from 'vue-router';
import {GameRound, GameRoundPlayer, Player} from '@/api/GameApi';

const router = useRouter();
const gameStore = useGameStore();
const playerStore = usePlayerStore();
const loading = ref(false);

const round = ref<GameRound>({
  code: '',
  name: '',
  game_type: '',
  start_time: new Date().toISOString(),
  players: [],
  version: 0
});

const players = ref(Array<Player>());
playerStore.allPlayers.then(v => players.value = v);
const gameTypes = computed(() => gameStore.gameTypes);
const isEditing = computed(() => !!round.value.code);

const selectedGameType = computed(() =>
    gameTypes.value.find(gt => gt.code === round.value.game_type)
);

const startRound = async () => {
  if (!round.value.game_type) return;

  loading.value = true;
  try {
    // Get all users and create players array
    const a_players: GameRoundPlayer[] = players.value.map(user => ({
      user_id: user.code,
      score: 0,
      is_moderator: false,
      team_name: undefined
    }));

    round.value.players = a_players;
  } catch (error) {
    console.error('Error starting round:', error);
  } finally {
    loading.value = false;
  }
};

const saveRound = async () => {
  try {
    const savedRound = isEditing.value
        ? await gameStore.updateRound(round.value)
        : await gameStore.addActiveRound(round.value);

    await router.push({
      name: 'GameRounds',
      params: { code: savedRound.code }
    });
  } catch (error) {
    console.error('Error saving game round:', error);
  }
};

onMounted(async () => {
  try {
    if (gameStore.gameTypes.length === 0) {
      await gameStore.loadGameTypes();
    }
  } catch (error) {
    console.error('Failed to load game types:', error);
  }
});
</script>