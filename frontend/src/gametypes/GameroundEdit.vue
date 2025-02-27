<template>
  <v-container>
    <v-row>
      <v-col>
        <h2>{{ isEditing ? 'Edit Game Round' : 'New Game Round' }}</h2>
        <v-form @submit.prevent="saveRound">
          <v-text-field
              v-model="round.name"
              label="Round Name"
              required
          />
          <v-select
              v-model="round.game_type"
              :items="gameTypes"
              item-title="name"
              item-value="code"
              label="Game Type"
              required
          />
          <v-list>
            <v-subheader>Players</v-subheader>
            <v-list-item v-for="(player, index) in round.players" :key="index">
              <v-row>
                <v-col>
                  <v-select
                      v-model="player.user_id"
                      :items="users"
                      item-title="name"
                      item-value="id"
                      label="Player"
                  />
                </v-col>
                <v-col>
                  <v-text-field
                      v-model.number="player.score"
                      label="Score"
                      type="number"
                  />
                </v-col>
                <v-col>
                  <v-checkbox
                      v-model="player.is_moderator"
                      label="Moderator"
                  />
                </v-col>
              </v-row>
            </v-list-item>
          </v-list>
          <v-btn type="submit" color="primary">Save</v-btn>
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue';
import { GameRoundView } from './types';

const isEditing = computed(() => !!round.value.code);

const round = ref<GameRound>({
  code: '',
  name: '',
  game_type: '',
  start_time: new Date().toISOString(),
  players: []
});

import { useGameStore } from '@/store/game';
import GameApi, {GameRound} from "@/api/GameApi";
const gameStore = useGameStore();

// When saving a round:
const saveRound = async () => {
  if (isEditing.value) {
    await gameStore.updateRound(round.value);
  } else {
    await gameStore.addActiveRound(round.value);
  }
  await router.push({ name: 'GameRounds' });
};
</script>