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
          <v-card class="mb-4" v-if="showLabels" >
            <v-card-title>Labels</v-card-title>
            <v-card-text>
              <div v-for="(label, index) in currentGameType.labels" :key="index" class="d-flex align-center mb-2">
                <v-text-field v-model="label.name" label="Label Name" class="mr-2" />
                <div class="color-box-wrapper mr-2">
                  <div
                      class="color-box"
                      :style="{ backgroundColor: label.color }"
                      @click="() => label.showPicker = !label.showPicker"
                  />
                  <v-menu
                      v-model="label.showPicker"
                      :close-on-content-click="false"
                      location="bottom"
                  >
                    <template v-slot:activator="{ props }">
                      <div v-bind="props"></div>
                    </template>
                    <v-card>
                      <v-color-picker
                          v-model="label.color"
                          hide-inputs
                          @update:model-value="() => label.showPicker = false"
                      />
                    </v-card>
                  </v-menu>
                </div>
                <v-text-field v-model="label.icon" label="Icon" class="mx-2" />
                <v-btn icon color="error" @click="removeLabel(index)">
                  <v-icon>mdi-delete</v-icon>
                </v-btn>
              </div>
              <v-btn color="primary" @click="addLabel">Add Label</v-btn>
            </v-card-text>
          </v-card>

          <!-- Teams Section -->
          <v-card class="mb-4" v-if="showTeams">
            <v-card-title>Teams</v-card-title>
            <v-card-text>
              <div v-for="(team, index) in currentGameType.teams" :key="index" class="d-flex align-center mb-2">
                <v-text-field v-model="team.name" label="Team Name" class="mr-2" />
                <div class="color-box-wrapper mr-2">
                  <div
                      class="color-box"
                      :style="{ backgroundColor: team.color }"
                      @click="() => team.showPicker = !team.showPicker"
                  />
                  <v-menu
                      v-model="team.showPicker"
                      :close-on-content-click="false"
                      location="bottom"
                  >
                    <template v-slot:activator="{ props }">
                      <div v-bind="props"></div>
                    </template>
                    <v-card>
                      <v-color-picker
                          v-model="team.color"
                          hide-inputs
                          @update:model-value="() => team.showPicker = false"
                      />
                    </v-card>
                  </v-menu>
                </div>
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
import {ref, onMounted, computed} from 'vue';
import GameApi, {GameType, Label, ScoringType, ScoringTypes} from '@/api/GameApi';

interface LabelUI extends Label {
  showPicker: boolean
}

interface GameTypeUI extends GameType {
  labels: Array<LabelUI>
  teams: Array<LabelUI>
}

const defaultGameType = {} as GameTypeUI;
const defaultLabel = { name: '', color: '#000000', icon: '', showPicker: false };
const defaultTeam = { name: '', color: '#000000', icon: '', showPicker: false };

const gameTypes = ref(Array<GameType>());
const currentGameType = ref( {...defaultGameType} as GameTypeUI);
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
  currentGameType.value = {...gameType} as GameTypeUI;
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

const showLabels = computed(() => {
  const labelScoringTypes: ScoringType[] = ['classic', 'custom'];
  return labelScoringTypes.includes(currentGameType.value.scoring_type);
});

const showTeams = computed(() => {
  const teamScoringTypes: ScoringType[] = ['mafia', 'custom', 'team_vs_team'];
  return teamScoringTypes.includes(currentGameType.value.scoring_type);
});

onMounted(fetchGameTypes);
</script>

<style scoped>
.color-box-wrapper {
  position: relative;
  width: 40px;
}

.color-box {
  width: 40px;
  height: 40px;
  border-radius: 4px;
  border: 1px solid #ccc;
  cursor: pointer;
}

.gap-2 {
  gap: 8px;
}
</style>