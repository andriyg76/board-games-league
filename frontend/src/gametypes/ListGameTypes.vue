<template>
  <n-grid :cols="24" :x-gap="16">
    <n-gi :span="24">
      <h2 style="font-size: 2rem; margin-bottom: 16px;">{{ t('gameTypes.title') }}</h2>
      <n-list>
        <n-list-item v-for="gameType in gameTypes" :key="gameType.code">
          <template #prefix>
            <n-icon v-if="gameType.icon" :size="24">
              <DiceIcon />
            </n-icon>
            <n-tag v-if="gameType.built_in" type="info" size="small" style="margin-left: 8px;">
              {{ t('gameTypes.builtIn') }}
            </n-tag>
          </template>
          <div>
            <div style="font-weight: 500;">
              {{ getGameTypeName(gameType) }} — {{ t(`scoring.${gameType.scoring_type}`) }}
            </div>
            <div style="font-size: 0.875rem; opacity: 0.7;">
              {{ t('gameTypes.players') }}: {{ gameType.min_players }}-{{ gameType.max_players }}
              <span v-if="gameType.roles?.length"> | {{ t('gameTypes.roles') }}: {{ gameType.roles.length }}</span>
            </div>
          </div>
          <template #suffix>
            <div style="display: flex; gap: 8px;">
              <n-button size="small" type="primary" @click="editGameType(gameType)">
                {{ t('gameTypes.edit') }}
              </n-button>
              <n-button 
                size="small"
                type="error"
                @click="deleteGameType(gameType.code)" 
                :disabled="gameType.built_in"
              >
                {{ t('gameTypes.delete') }}
              </n-button>
            </div>
          </template>
        </n-list-item>
      </n-list>
    </n-gi>
  </n-grid>

  <n-divider style="margin: 24px 0;" />

  <n-grid :cols="24" :x-gap="16">
    <n-gi :span="24">
      <h3 style="font-size: 1.5rem; margin-bottom: 16px;">{{ isEditing ? t('gameTypes.edit') : t('gameTypes.create') }} {{ t('gameTypes.gameType') }}</h3>
      <n-form @submit.prevent="saveGameType">
        <!-- Key -->
        <n-form-item :label="t('gameTypes.key')" required>
          <n-input 
            v-model:value="currentGameType.key" 
            :disabled="isEditing && currentGameType.built_in"
          />
        </n-form-item>

        <!-- Localized Names -->
        <n-card style="margin-bottom: 16px;">
          <template #header>
            {{ t('gameTypes.localizedNames') }}
          </template>
          <n-form-item label="English" required>
            <n-input v-model:value="currentGameType.names.en" />
          </n-form-item>
          <n-form-item label="Українська">
            <n-input v-model:value="currentGameType.names.uk" />
          </n-form-item>
          <n-form-item label="Eesti">
            <n-input v-model:value="currentGameType.names.et" />
          </n-form-item>
        </n-card>

        <!-- Icon -->
        <n-form-item :label="t('gameTypes.icon')">
          <n-input 
            v-model:value="currentGameType.icon" 
            placeholder="mdi-cards-playing"
          />
        </n-form-item>

        <!-- Scoring Type -->
        <n-form-item :label="t('gameTypes.scoringType')" required>
          <n-select
            v-model:value="currentGameType.scoring_type"
            :options="Object.keys(ScoringTypes).map(key => ({ label: t(`scoring.${key}`), value: key }))"
          />
        </n-form-item>

        <!-- Min/Max Players -->
        <n-grid :cols="24" :x-gap="8">
          <n-gi :span="24" :responsive="{ m: 12 }">
            <n-form-item :label="t('gameTypes.minPlayers')">
              <n-input-number 
                v-model:value="currentGameType.min_players" 
                :min="1"
                style="width: 100%;"
              />
            </n-form-item>
          </n-gi>
          <n-gi :span="24" :responsive="{ m: 12 }">
            <n-form-item :label="t('gameTypes.maxPlayers')">
              <n-input-number 
                v-model:value="currentGameType.max_players" 
                :min="1"
                style="width: 100%;"
              />
            </n-form-item>
          </n-gi>
        </n-grid>

        <!-- Roles Section -->
        <n-card style="margin-bottom: 16px;">
          <template #header>
            {{ t('gameTypes.roles') }}
          </template>
          <div v-for="(role, index) in currentGameType.roles" :key="index" class="role-item" style="margin-bottom: 16px; padding: 16px; border: 1px solid rgba(0, 0, 0, 0.12); border-radius: 4px; background-color: rgba(0, 0, 0, 0.02);">
            <n-grid :cols="24" :x-gap="8">
              <n-gi :span="24" :responsive="{ m: 6 }">
                <n-form-item :label="t('gameTypes.roleKey')">
                  <n-input v-model:value="role.key" size="small" />
                </n-form-item>
              </n-gi>
              <n-gi :span="24" :responsive="{ m: 6 }">
                <n-form-item label="Name (EN)">
                  <n-input v-model:value="role.names.en" size="small" />
                </n-form-item>
              </n-gi>
              <n-gi :span="24" :responsive="{ m: 6 }">
                <n-form-item label="Назва (UK)">
                  <n-input v-model:value="role.names.uk" size="small" />
                </n-form-item>
              </n-gi>
              <n-gi :span="24" :responsive="{ m: 6 }">
                <n-form-item :label="t('gameTypes.roleType')">
                  <n-select
                    v-model:value="role.role_type"
                    :options="roleTypeOptions.map(opt => ({ label: t(`roleTypes.${opt.value}`), value: opt.value }))"
                    size="small"
                  />
                </n-form-item>
              </n-gi>
            </n-grid>
            <n-grid :cols="24" :x-gap="8">
              <n-gi :span="24" :responsive="{ m: 8 }">
                <div style="display: flex; align-items: center; gap: 8px;">
                  <div 
                    class="color-box"
                    :style="{ backgroundColor: role.color }"
                    @click="() => role.showPicker = !role.showPicker"
                  />
                  <n-input 
                    v-model:value="role.color" 
                    :placeholder="t('gameTypes.color')"
                    size="small"
                    readonly
                    @click="() => role.showPicker = !role.showPicker"
                    style="flex: 1;"
                  />
                  <n-color-picker
                    v-if="role.showPicker"
                    v-model:value="role.color"
                    :show-alpha="false"
                    @update:value="() => role.showPicker = false"
                    style="position: absolute; z-index: 1000;"
                  />
                </div>
              </n-gi>
              <n-gi :span="24" :responsive="{ m: 8 }">
                <n-form-item :label="t('gameTypes.icon')">
                  <n-input v-model:value="role.icon" size="small" />
                </n-form-item>
              </n-gi>
              <n-gi :span="24" :responsive="{ m: 8 }" style="display: flex; align-items: center;">
                <n-button type="error" size="small" @click="removeRole(index)">
                  <template #icon>
                    <n-icon><DeleteIcon /></n-icon>
                  </template>
                </n-button>
              </n-gi>
            </n-grid>
          </div>
          <n-button type="primary" @click="addRole">
            <template #icon>
              <n-icon><AddIcon /></n-icon>
            </template>
            {{ t('gameTypes.addRole') }}
          </n-button>
        </n-card>

        <div style="display: flex; gap: 8px;">
          <n-button type="success" @click="saveGameType">
            {{ isEditing ? t('gameTypes.update') : t('gameTypes.create') }}
          </n-button>
          <n-button type="error" @click="cancelEdit">
            {{ t('gameTypes.cancel') }}
          </n-button>
        </div>
      </n-form>
    </n-gi>
  </n-grid>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { NGrid, NGi, NList, NListItem, NIcon, NTag, NButton, NDivider, NForm, NFormItem, NInput, NCard, NSelect, NInputNumber, NColorPicker } from 'naive-ui';
import { Add as AddIcon, Dice as DiceIcon, Trash as DeleteIcon } from '@vicons/ionicons5';
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
</style>
