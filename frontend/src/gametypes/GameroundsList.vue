<template>
  <v-container>
    <v-row>
      <v-col>
        <h2>Game Rounds</h2>
        <v-list>
          <v-list-item v-for="round in gameRounds" :key="round.id">
            <v-list-item-content>
              <v-list-item-title>{{ round.name }}</v-list-item-title>
              <v-list-item-subtitle>
                Started: {{ formatDate(round.start_time) }}
                {{ round.end_time ? `| Ended: ${formatDate(round.end_time)}` : '' }}
              </v-list-item-subtitle>
            </v-list-item-content>
            <v-list-item-action>
              <v-btn @click="editRound(round)" color="primary">Edit</v-btn>
              <v-btn v-if="!round.end_time" @click="finalizeRound(round.id)" color="success">Finalize</v-btn>
            </v-list-item-action>
          </v-list-item>
        </v-list>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { GameRoundView } from './types';
import {useRouter} from "vue-router";

const gameRounds = ref<GameRoundView[]>([]);

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString();
};

const finalizeRound = (id: string) => {
  // Implement finalization logic
};

onMounted(async () => {
  // Fetch game rounds
});

const router = useRouter();

const editRound = (round: GameRoundView) => {
  router.push({ name: 'EditGameRound', params: { id: round.code }});
};

const createNewRound = () => {
  router.push({ name: 'NewGameRound' });
};
</script>