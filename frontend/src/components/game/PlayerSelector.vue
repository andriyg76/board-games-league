<template>
  <n-grid :cols="24" :x-gap="16">
    <!-- Left panel: Available players -->
    <n-gi :span="24" :responsive="{ m: 12 }">
      <n-card>
        <template #header>
          <div style="display: flex; align-items: center; justify-content: space-between;">
            <span>{{ $t('game.availablePlayers') }}</span>
            <n-button
              size="small"
              quaternary
              type="primary"
              @click="showCreateDialog = true"
            >
              <template #icon>
                <n-icon><AddIcon /></n-icon>
              </template>
              {{ $t('game.addVirtual') }}
            </n-button>
          </div>
        </template>
        
        <n-input
          v-model:value="searchQuery"
          :placeholder="$t('common.search')"
          clearable
          style="margin-bottom: 16px;"
        >
          <template #prefix>
            <n-icon><SearchIcon /></n-icon>
          </template>
        </n-input>
        
        <n-list style="max-height: 400px; overflow-y: auto;">
          <n-list-item
            v-for="player in filteredAvailablePlayers"
            :key="player.membership_id"
            clickable
            :style="{ opacity: player.is_virtual ? 0.8 : 1 }"
            @click="addPlayer(player)"
          >
            <template #prefix>
              <n-avatar :size="32" round>
                <img v-if="player.avatar" :src="player.avatar" />
                <template v-else>
                  <n-icon><PersonIcon /></n-icon>
                </template>
              </n-avatar>
            </template>
            
            <div>
              <div style="display: flex; align-items: center; gap: 8px;">
                <span style="font-weight: 500;">{{ player.alias }}</span>
                <n-tag v-if="player.is_virtual" size="small" type="info">
                  {{ $t('game.virtual') }}
                </n-tag>
              </div>
            </div>
            
            <template #suffix>
              <n-button
                quaternary
                circle
                size="small"
                type="primary"
              >
                <template #icon>
                  <n-icon><AddIcon /></n-icon>
                </template>
              </n-button>
            </template>
          </n-list-item>
          
          <n-list-item v-if="filteredAvailablePlayers.length === 0">
            <div style="opacity: 0.7;">
              {{ $t('game.noPlayersFound') }}
            </div>
          </n-list-item>
        </n-list>
      </n-card>
    </n-gi>
    
    <!-- Right panel: Selected players -->
    <n-gi :span="24" :responsive="{ m: 12 }">
      <n-card>
        <template #header>
          <div style="display: flex; align-items: center; gap: 8px;">
            <span>{{ $t('game.selectedPlayers') }}</span>
            <n-tag 
              size="small" 
              :type="playerCountTagType"
            >
              {{ selectedPlayers.length }} / {{ maxPlayers }}
            </n-tag>
          </div>
        </template>
        
        <n-alert
          v-if="selectedPlayers.length < minPlayers"
          type="warning"
          size="small"
          style="margin-bottom: 16px;"
        >
          {{ $t('game.minPlayersWarning', { min: minPlayers }) }}
        </n-alert>
        
        <n-list style="max-height: 400px; overflow-y: auto;">
          <n-list-item
            v-for="(player, index) in selectedPlayers"
            :key="player.membership_id"
            clickable
            :style="getPlayerItemStyle(player)"
            @click="toggleModerator(player)"
          >
            <template #prefix>
              <n-avatar :size="32" round>
                <img v-if="player.avatar" :src="player.avatar" />
                <template v-else>
                  <n-icon><PersonIcon /></n-icon>
                </template>
              </n-avatar>
            </template>
            
            <div>
              <div style="display: flex; align-items: center; gap: 8px; flex-wrap: wrap;">
                <span style="font-weight: 500;">{{ player.alias }}</span>
                <n-tag 
                  v-if="player.membership_id === currentPlayerMembershipId" 
                  size="small" 
                  type="primary"
                >
                  {{ $t('game.you') }}
                </n-tag>
                <n-tag v-if="player.is_virtual" size="small" type="info">
                  {{ $t('game.virtual') }}
                </n-tag>
                <n-icon 
                  v-if="hasModerator && player.membership_id === selectedModeratorId" 
                  size="16" 
                  color="#f0a020"
                >
                  <CrownIcon />
                </n-icon>
              </div>
            </div>
            
            <template #suffix>
              <n-button
                quaternary
                circle
                size="small"
                type="error"
                @click.stop="removePlayer(index)"
              >
                <template #icon>
                  <n-icon><CloseIcon /></n-icon>
                </template>
              </n-button>
            </template>
          </n-list-item>
          
          <n-list-item v-if="selectedPlayers.length === 0">
            <div style="opacity: 0.7;">
              {{ $t('game.noPlayersSelected') }}
            </div>
          </n-list-item>
        </n-list>
        
        <div v-if="hasModerator" style="margin-top: 16px; font-size: 0.875rem; opacity: 0.7; display: flex; align-items: center; gap: 4px;">
          <n-icon size="16"><InformationIcon /></n-icon>
          {{ $t('game.clickToSelectModerator') }}
        </div>
      </n-card>
    </n-gi>
  </n-grid>
  
  <!-- Create virtual player dialog -->
  <CreateVirtualPlayerDialog
    v-model="showCreateDialog"
    :league-code="leagueCode"
    @created="onVirtualPlayerCreated"
  />
</template>

<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
import { NGrid, NGi, NCard, NList, NListItem, NAvatar, NIcon, NButton, NTag, NInput, NAlert } from 'naive-ui';
import { 
  Add as AddIcon,
  Search as SearchIcon,
  Person as PersonIcon,
  Close as CloseIcon,
  Star as CrownIcon,
  Information as InformationIcon
} from '@vicons/ionicons5';
import { SuggestedPlayer, SuggestedPlayersResponse } from '@/api/LeagueApi';
import CreateVirtualPlayerDialog from './CreateVirtualPlayerDialog.vue';

const props = defineProps<{
  suggestedPlayers: SuggestedPlayersResponse | null;
  minPlayers: number;
  maxPlayers: number;
  hasModerator: boolean;
  leagueCode: string;
}>();

const emit = defineEmits<{
  'update:selectedPlayers': [players: SuggestedPlayer[]];
  'update:moderatorId': [id: string | null];
}>();

const searchQuery = ref('');
const showCreateDialog = ref(false);
const selectedPlayers = ref<SuggestedPlayer[]>([]);
const selectedModeratorId = ref<string | null>(null);

// Current player's membership ID for highlighting
const currentPlayerMembershipId = computed(() => 
  props.suggestedPlayers?.current_player?.membership_id || null
);

// Combine all available players (recent + other, excluding already selected)
const allAvailablePlayers = computed(() => {
  if (!props.suggestedPlayers) return [];
  
  const selectedIds = new Set(selectedPlayers.value.map(p => p.membership_id));
  const players: SuggestedPlayer[] = [];
  
  // Add recent players first
  for (const player of props.suggestedPlayers.recent_players) {
    if (!selectedIds.has(player.membership_id)) {
      players.push(player);
    }
  }
  
  // Then add other players
  for (const player of props.suggestedPlayers.other_players) {
    if (!selectedIds.has(player.membership_id)) {
      players.push(player);
    }
  }
  
  return players;
});

// Filter by search query
const filteredAvailablePlayers = computed(() => {
  if (!searchQuery.value) return allAvailablePlayers.value;
  
  const query = searchQuery.value.toLowerCase();
  return allAvailablePlayers.value.filter(p => 
    p.alias.toLowerCase().includes(query)
  );
});

// Player count tag type based on min/max
const playerCountTagType = computed<'success' | 'warning' | 'error' | 'info'>(() => {
  const count = selectedPlayers.value.length;
  if (count < props.minPlayers) return 'warning';
  if (count > props.maxPlayers) return 'error';
  return 'success';
});

// Get style for player item
const getPlayerItemStyle = (player: SuggestedPlayer) => {
  const style: Record<string, any> = {
    opacity: player.is_virtual ? 0.8 : 1,
  };
  
  if (props.hasModerator && player.membership_id === selectedModeratorId.value) {
    style.backgroundColor = 'rgba(240, 160, 32, 0.1)';
    style.borderLeft = '3px solid #f0a020';
  }
  
  return style;
};

// Add player to selection
const addPlayer = (player: SuggestedPlayer) => {
  if (selectedPlayers.value.length >= props.maxPlayers) return;
  if (selectedPlayers.value.find(p => p.membership_id === player.membership_id)) return;
  
  selectedPlayers.value.push(player);
  emit('update:selectedPlayers', selectedPlayers.value);
};

// Remove player from selection
const removePlayer = (index: number) => {
  const removed = selectedPlayers.value[index];
  selectedPlayers.value.splice(index, 1);
  
  // Clear moderator if removed
  if (removed.membership_id === selectedModeratorId.value) {
    selectedModeratorId.value = null;
    emit('update:moderatorId', null);
  }
  
  emit('update:selectedPlayers', selectedPlayers.value);
};

// Toggle moderator selection
const toggleModerator = (player: SuggestedPlayer) => {
  if (!props.hasModerator) return;
  
  if (selectedModeratorId.value === player.membership_id) {
    selectedModeratorId.value = null;
  } else {
    selectedModeratorId.value = player.membership_id;
  }
  emit('update:moderatorId', selectedModeratorId.value);
};

// Handle virtual player creation
const onVirtualPlayerCreated = (player: SuggestedPlayer) => {
  // Add to selected players
  addPlayer(player);
};

// Auto-fill players when suggested players change
watch(() => props.suggestedPlayers, (newValue) => {
  if (!newValue || selectedPlayers.value.length > 0) return;
  
  // Auto-fill with current player first
  if (newValue.current_player) {
    selectedPlayers.value.push(newValue.current_player);
  }
  
  // Then fill with recent players
  for (const player of newValue.recent_players) {
    if (selectedPlayers.value.length >= props.maxPlayers) break;
    if (!selectedPlayers.value.find(p => p.membership_id === player.membership_id)) {
      selectedPlayers.value.push(player);
    }
  }
  
  // Then fill with other players
  for (const player of newValue.other_players) {
    if (selectedPlayers.value.length >= props.maxPlayers) break;
    if (!selectedPlayers.value.find(p => p.membership_id === player.membership_id)) {
      selectedPlayers.value.push(player);
    }
  }
  
  emit('update:selectedPlayers', selectedPlayers.value);
}, { immediate: true });

// Expose methods for parent component
defineExpose({
  getSelectedPlayers: () => selectedPlayers.value,
  getModeratorId: () => selectedModeratorId.value,
});
</script>
