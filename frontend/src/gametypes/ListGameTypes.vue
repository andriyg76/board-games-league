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
        <v-form>
          <v-text-field v-model="currentGameType.name" label="Game Type Name" />
          <v-select
              v-model="currentGameType.scoring_type"
              :items="Object.keys(ScoringTypes)"
              :item-title="(item) => ScoringTypes[item as ScoringType]"
              :item-value="(item) => item"
              label="Scoring Type"
              required
          />

          <!-- Labels Section -->
          <v-card class="mb-4">
            <v-card-title>Labels</v-card-title>
            <v-card-text>
              <div v-for="(label, index) in currentGameType.labels" :key="index" class="d-flex align-center mb-2">
                <v-text-field v-model="label.name" label="Label Name" class="mr-2" />
                <v-color-picker v-model="label.color" hide-inputs />
                <v-text-field v-model="label.icon" label="Icon" class="mx-2" />
                <v-btn icon color="error" @click="removeLabel(index)">
                  <v-icon>mdi-delete</v-icon>
                </v-btn>
              </div>
              <v-btn color="primary" @click="addLabel">Add Label</v-btn>
            </v-card-text>
          </v-card>

          <!-- Teams Section -->
          <v-card class="mb-4">
            <v-card-title>Teams</v-card-title>
            <v-card-text>
              <div v-for="(team, index) in currentGameType.teams" :key="index" class="d-flex align-center mb-2">
                <v-text-field v-model="team.name" label="Team Name" class="mr-2" />
                <v-color-picker v-model="team.color" hide-inputs class="v-color-picker" />
                <v-text-field v-model="team.icon" label="Icon" class="mx-2" />
                <v-btn icon color="error" @click="removeTeam(index)">
                  <v-icon>mdi-delete</v-icon>
                </v-btn>
              </div>
              <v-btn color="primary" @click="addTeam">Add Team</v-btn>
            </v-card-text>
          </v-card>


          <div class="d-flex gap-2">
            <v-btn @click="saveGameType" color="success">{{ isEditing ? 'Update' : 'Create' }}</v-btn>
            <v-btn @click="cancelEdit" color="error">Cancel</v-btn>
          </div>
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import GameApi, {GameType, Label, ScoringType, ScoringTypes} from '@/api/GameApi';

const defaultGameType = {} as GameType;
const defaultLabel = { name: '', color: '#000000', icon: '' };
const defaultTeam = { name: '', color: '#000000', icon: '' };

const gameTypes = ref(Array<GameType>());
const currentGameType = ref({...defaultGameType});
const isEditing = ref(false);

const cancelEdit = () => {
  currentGameType.value = {...defaultGameType};
  isEditing.value = false;
};

const fetchGameTypes = async () => {
  try {
    gameTypes.value = await GameApi.getGameTypes();
  } catch (error) {
    console.error('Error fetching game types:', error);
  }

};
const saveGameType = async () => {
  try {
    if (isEditing.value) {
      await GameApi.updateGameType(currentGameType.value.code, currentGameType.value);
    } else {
      await GameApi.createGameType(currentGameType.value);
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
    await GameApi.deleteGameType(code);
    await fetchGameTypes();
  } catch (error) {
    console.error('Error deleting game type:', error);
  }
};

const addLabel = () => {
  if (!currentGameType.value.labels) {
    currentGameType.value.labels = [];
  }
  currentGameType.value.labels.push({ ...defaultLabel });
};

const removeLabel = (index: number) => {
  currentGameType.value.labels.splice(index, 1);
};

const addTeam = () => {
  if (!currentGameType.value.teams) {
    currentGameType.value.teams = [];
  }
  currentGameType.value.teams.push({ ...defaultTeam });
};

const removeTeam = (index: number) => {
  currentGameType.value.teams.splice(index, 1);
};

onMounted(fetchGameTypes);
</script>

<style scoped>
.v-color-picker {
  max-width: 100px;
}

.gap-2 {
  gap: 8px;
}
</style>