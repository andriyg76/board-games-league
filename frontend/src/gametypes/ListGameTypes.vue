<template>
  <v-container>
    <v-row>
      <v-col>
        <h2>{{ t('gameTypes.title') }}</h2>
        <v-list>
          <v-list-item v-for="gameType in gameTypes" :key="gameType.code">
            <template v-slot:prepend>
              <v-icon v-if="gameType.icon">{{ gameType.icon }}</v-icon>
              <v-chip v-if="gameType.built_in" size="x-small" color="info" class="ml-2">
                {{ t('gameTypes.builtIn') }}
              </v-chip>
            </template>
            <v-list-item-title>
              {{ getGameTypeName(gameType) }} — {{ t(`scoring.${gameType.scoring_type}`) }}
            </v-list-item-title>
            <v-list-item-subtitle>
              {{ t('gameTypes.players') }}: {{ gameType.min_players }}-{{ gameType.max_players }}
              <span v-if="gameType.roles?.length"> | {{ t('gameTypes.roles') }}: {{ gameType.roles.length }}</span>
            </v-list-item-subtitle>
            <template v-slot:append>
              <v-btn @click="editGameType(gameType)" color="primary" size="small" class="mr-2">
                {{ t('gameTypes.edit') }}
              </v-btn>
              <v-btn 
                @click="deleteGameType(gameType.code)" 
                color="error" 
                size="small"
                :disabled="gameType.built_in"
              >
                {{ t('gameTypes.delete') }}
              </v-btn>
            </template>
          </v-list-item>
        </v-list>
      </v-col>
    </v-row>

    <v-divider class="my-4" />

    <v-row>
      <v-col>
        <h3>{{ isEditing ? t('gameTypes.edit') : t('gameTypes.create') }} {{ t('gameTypes.gameType') }}</h3>
        <v-form @submit.prevent="saveGameType">
          <!-- Key -->
          <v-text-field 
            v-model="currentGameType.key" 
            :label="t('gameTypes.key')"
            :disabled="isEditing && currentGameType.built_in"
            required
          />

          <!-- Localized Names -->
          <v-card class="mb-4">
            <v-card-title>{{ t('gameTypes.localizedNames') }}</v-card-title>
            <v-card-text>
              <v-text-field 
                v-model="currentGameType.names.en" 
                label="English"
                required
              />
              <v-text-field 
                v-model="currentGameType.names.uk" 
                label="Українська"
              />
              <v-text-field 
                v-model="currentGameType.names.et" 
                label="Eesti"
              />
            </v-card-text>
          </v-card>

          <!-- Icon -->
          <v-text-field 
            v-model="currentGameType.icon" 
            :label="t('gameTypes.icon')"
            placeholder="mdi-cards-playing"
          />

          <!-- Scoring Type -->
          <v-select
            v-model="currentGameType.scoring_type"
            :items="Object.keys(ScoringTypes)"
            :item-title="(item) => t(`scoring.${item}`)"
            :item-value="(item) => item"
            :label="t('gameTypes.scoringType')"
            required
          />

          <!-- Min/Max Players -->
          <v-row>
            <v-col cols="6">
              <v-text-field 
                v-model.number="currentGameType.min_players" 
                :label="t('gameTypes.minPlayers')"
                type="number"
                min="1"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field 
                v-model.number="currentGameType.max_players" 
                :label="t('gameTypes.maxPlayers')"
                type="number"
                min="1"
              />
            </v-col>
          </v-row>

          <!-- Roles Section -->
          <v-card class="mb-4">
            <v-card-title>{{ t('gameTypes.roles') }}</v-card-title>
            <v-card-text>
              <div v-for="(role, index) in currentGameType.roles" :key="index" class="role-item mb-4 pa-3 border rounded">
                <v-row>
                  <v-col cols="12" md="3">
                    <v-text-field 
                      v-model="role.key" 
                      :label="t('gameTypes.roleKey')"
                      dense
                    />
                  </v-col>
                  <v-col cols="12" md="3">
                    <v-text-field 
                      v-model="role.names.en" 
                      label="Name (EN)"
                      dense
                    />
                  </v-col>
                  <v-col cols="12" md="3">
                    <v-text-field 
                      v-model="role.names.uk" 
                      label="Назва (UK)"
                      dense
                    />
                  </v-col>
                  <v-col cols="12" md="3">
                    <v-select
                      v-model="role.role_type"
                      :items="roleTypeOptions"
                      :item-title="(item) => t(`roleTypes.${item.value}`)"
                      item-value="value"
                      :label="t('gameTypes.roleType')"
                      dense
                    />
                  </v-col>
                </v-row>
                <v-row>
                  <v-col cols="12" md="4">
                    <div class="d-flex align-center">
                      <div 
                        class="color-box mr-2"
                        :style="{ backgroundColor: role.color }"
                        @click="() => role.showPicker = !role.showPicker"
                      />
                      <v-menu
                        v-model="role.showPicker"
                        :close-on-content-click="false"
                        location="bottom"
                      >
                        <template v-slot:activator="{ props }">
                          <v-text-field 
                            v-model="role.color" 
                            :label="t('gameTypes.color')"
                            v-bind="props"
                            dense
                            readonly
                          />
                        </template>
                        <v-card>
                          <v-color-picker
                            v-model="role.color"
                            hide-inputs
                            @update:model-value="() => role.showPicker = false"
                          />
                        </v-card>
                      </v-menu>
                    </div>
                  </v-col>
                  <v-col cols="12" md="4">
                    <v-text-field 
                      v-model="role.icon" 
                      :label="t('gameTypes.icon')"
                      dense
                    />
                  </v-col>
                  <v-col cols="12" md="4" class="d-flex align-center">
                    <v-btn icon color="error" @click="removeRole(index)" size="small">
                      <v-icon>mdi-delete</v-icon>
                    </v-btn>
                  </v-col>
                </v-row>
              </div>
              <v-btn color="primary" @click="addRole" prepend-icon="mdi-plus">
                {{ t('gameTypes.addRole') }}
              </v-btn>
            </v-card-text>
          </v-card>

          <div class="d-flex gap-2">
            <v-btn type="submit" color="success">
              {{ isEditing ? t('gameTypes.update') : t('gameTypes.create') }}
            </v-btn>
            <v-btn @click="cancelEdit" color="error">
              {{ t('gameTypes.cancel') }}
            </v-btn>
          </div>
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import GameApi, { GameType, Role, ScoringTypes, getLocalizedName } from '@/api/GameApi';
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()

interface RoleUI extends Role {
  showPicker: boolean
}

interface GameTypeUI extends Omit<GameType, 'roles'> {
  roles: RoleUI[]
}

const defaultGameType: GameTypeUI = {
  code: '',
  key: '',
  names: { en: '', uk: '', et: '' },
  icon: '',
  scoring_type: 'classic',
  roles: [],
  min_players: 2,
  max_players: 6,
  built_in: false,
  version: 0
};

const defaultRole: RoleUI = {
  key: '',
  names: { en: '', uk: '' },
  color: '#4CAF50',
  icon: '',
  role_type: 'optional_one',
  showPicker: false
};

const gameTypes = ref<GameType[]>([]);
const currentGameType = ref<GameTypeUI>({ ...defaultGameType, names: { ...defaultGameType.names }, roles: [] });
const isEditing = ref(false);

const roleTypeOptions = [
  { value: 'optional' },
  { value: 'optional_one' },
  { value: 'exactly_one' },
  { value: 'required' },
  { value: 'multiple' },
  { value: 'moderator' },
];

const getGameTypeName = (gameType: GameType): string => {
  return getLocalizedName(gameType.names, locale.value);
};

const cancelEdit = () => {
  currentGameType.value = { ...defaultGameType, names: { ...defaultGameType.names }, roles: [] };
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
    // Clean up roles - remove showPicker property
    const cleanRoles = currentGameType.value.roles.map(({ showPicker: _showPicker, ...role }) => role);
    const gameTypeToSave = {
      ...currentGameType.value,
      roles: cleanRoles
    };

    if (isEditing.value) {
      await GameApi.updateGameType(currentGameType.value.code, gameTypeToSave);
    } else {
      await GameApi.createGameType(gameTypeToSave);
    }
    await fetchGameTypes();
    cancelEdit();
  } catch (error) {
    console.error('Error saving game type:', error);
  }
};

const editGameType = (gameType: GameType) => {
  const rolesWithPicker: RoleUI[] = (gameType.roles || []).map(role => ({
    ...role,
    names: { ...role.names },
    showPicker: false
  }));
  
  currentGameType.value = {
    ...gameType,
    names: { ...gameType.names },
    roles: rolesWithPicker
  };
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

const addRole = () => {
  currentGameType.value.roles.push({ 
    ...defaultRole, 
    names: { ...defaultRole.names },
    showPicker: false 
  });
};

const removeRole = (index: number) => {
  currentGameType.value.roles.splice(index, 1);
};

onMounted(fetchGameTypes);
</script>

<style scoped>
.color-box {
  width: 40px;
  height: 40px;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid #ccc;
}

.role-item {
  background-color: rgba(0, 0, 0, 0.02);
}

.border {
  border: 1px solid rgba(0, 0, 0, 0.12);
}
</style>
