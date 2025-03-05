<template>
  <v-container>
    <v-row>
      <v-col>
        <h2>{{ isEditing ? 'Edit Game Round' : 'New Game Round' }}</h2>
        <v-form @submit.prevent="saveRound">
          <v-select
              v-model="round.game_type"
              :items="gameTypes"
              item-title="name"
              item-value="code"
              label="Game Type"
              required
              :disabled="isEditing"
          />
          <v-text-field
              v-model="round.name"
              label="Round Name"
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
import {ref, computed, onMounted} from 'vue';
import { GameRoundView } from './types';

const round = ref<GameRound>({
  code: '',
  name: '',
  game_type: '',
  start_time: new Date().toISOString(),
  players: [],
  version: 0
});


const isEditing = computed(() => !!round.value.code);
const gameTypes = computed(() => gameStore.gameTypes);

import { useGameStore } from '@/store/game';
import {GameRound} from "@/api/GameApi";
import {useRouter} from "vue-router";
const gameStore = useGameStore();
const  router = useRouter();

// When saving a round:
const saveRound = async () => {
  let savedRound;
  if (isEditing.value) {
    savedRound = await gameStore.addActiveRound(round.value);
  } else {
    savedRound = await gameStore.addActiveRound(round.value);
  }
  await router.push({ name: 'GameRounds',
    params: { code: savedRound.code }});
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