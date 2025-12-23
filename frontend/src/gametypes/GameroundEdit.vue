<template>
  <v-container>
    <v-row>
      <v-col>
        <h2>{{ isEditing ? 'Edit Game Round' : 'New Game Round' }}</h2>

        <div v-if="loadingRound" class="text-center pa-4">
          <v-progress-circular indeterminate color="primary" size="64" />
          <p class="mt-4">Loading game round...</p>
        </div>

        <v-form v-else @submit.prevent="saveRound">
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
import { GameRound, GameRoundPlayer, Player } from '@/api/GameApi';
import GameApi from '@/api/GameApi';

const props = defineProps<{
  id?: string;
}>();

const router = useRouter();
const gameStore = useGameStore();
const playerStore = usePlayerStore();
const loading = ref(false);
const loadingRound = ref(false);

const round = ref<GameRound>({
  code: '',
  name: '',
  game_type: '',
  start_time: new Date().toISOString(),
  players: [],
  version: 0
});

const players = computed(() => {
  const playerList = playerStore.players || [];
  return playerList.map((p: Player) => ({
    code: p.code,
    alias: p.alias,
    title: p.alias,
    props: {
      avatar: p.avatar,
      prependAvatar: p.avatar,
    }
  }));
});

const gameTypes = computed(() => gameStore.gameTypes);
const isEditing = computed(() => !!round.value.code);

const selectedGameType = computed(() =>
    gameTypes.value.find(gt => gt.code === round.value.game_type)
);

const loadGameRound = async () => {
  if (!props.id) return;

  loadingRound.value = true;
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
    loadingRound.value = false;
  }
};

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
    // Load game types first
    if (gameStore.gameTypes.length === 0) {
      await gameStore.loadGameTypes();
    }

    // Load players
    if (!playerStore.players || playerStore.players.length === 0) {
      await GameApi.listPlayers().then(p => {
        playerStore.players = p;
      });
    }

    // If editing, load the game round
    if (props.id) {
      await loadGameRound();
    }
  } catch (error) {
    console.error('Failed to load data:', error);
  }
});
</script>