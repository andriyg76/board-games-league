<template>
  <v-container>
    <v-row>
      <v-col>
        <h2>Game Types</h2>
        <v-list>
          <v-list-item v-for="gameType in gameTypes" :key="gameType.code">
            <v-list-item-content>
              <v-list-item-title>{{ gameType.name }} -- {{ ScoringTypes[gameType.scoring_type] }}</v-list-item-title>
            </v-list-item-content>
            <v-list-item-action>
              <v-btn @click="editGameType(gameType)" color="primary">Edit</v-btn>
              <v-btn @click="deleteGameType(gameType.code)" color="error">Delete</v-btn>
            </v-list-item-action>
          </v-list-item>
        </v-list>
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <h3>{{ isEditing ? 'Edit' : 'Create' }} Game Type</h3>
        <v-text-field v-model="currentGameType.name" label="Game Type Name" />
        <v-select
            v-model="currentGameType.scoring_type"
            :items="Object.keys(ScoringTypes)"
            :item-title="(item) => ScoringTypes[item as ScoringType]"
            :item-value="(item) => item"
            label="Scoring Type"
            required
        />
        <v-btn @click="saveGameType" color="success">{{ isEditing ? 'Update' : 'Create' }}</v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import GameTypeApi, {GameType, ScoringType, ScoringTypes } from '@/api/GameApi';

const defaultGameType = {icon: "", labels: [], maxPlayers: 6, minPlayers: 1, teams: [], version: 0, code: '', name: '', scoring_type: 'classic' };

const gameTypes = ref(Array<GameType>());
const currentGameType = ref({...defaultGameType} as GameType);
const isEditing = ref(false);

const fetchGameTypes = async () => {
  try {
    gameTypes.value = await GameTypeApi.getGameTypes();
  } catch (error) {
    console.error('Error fetching game types:', error);
  }

};
const saveGameType = async () => {
  try {
    if (isEditing.value) {
      await GameTypeApi.updateGameType(currentGameType.value.code, currentGameType.value);
    } else {
      await GameTypeApi.createGameType(currentGameType.value);
    }
    await fetchGameTypes();
    currentGameType.value = {...defaultGameType};
    isEditing.value = false;
  } catch (error) {
    console.error('Error saving game type:', error);
  }
};

const editGameType = (gameType: GameType) => {
  currentGameType.value = { ...gameType };
  isEditing.value = true;
};

const deleteGameType = async (code: string) => {
  try {
    await GameTypeApi.deleteGameType(code);
    await fetchGameTypes();
  } catch (error) {
    console.error('Error deleting game type:', error);
  }
};

onMounted(fetchGameTypes);
</script>
