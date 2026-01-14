<template>
  <v-row>
    <!-- Left panel: Available players -->
    <v-col cols="12" md="6">
      <v-card>
        <v-card-title class="d-flex align-center">
          <span>{{ $t('game.availablePlayers') }}</span>
          <v-spacer />
          <v-btn
            size="small"
            variant="tonal"
            color="primary"
            @click="showCreateDialog = true"
          >
            <v-icon start>mdi-plus</v-icon>
            {{ $t('game.addVirtual') }}
          </v-btn>
        </v-card-title>
        
        <v-card-text>
          <v-text-field
            v-model="searchQuery"
            :label="$t('common.search')"
            prepend-inner-icon="mdi-magnify"
            variant="outlined"
            density="compact"
            clearable
            class="mb-2"
          />
          
          <v-list density="compact" class="available-players-list">
            <v-list-item
              v-for="player in filteredAvailablePlayers"
              :key="player.membership_id"
              :class="{ 'virtual-player': player.is_virtual }"
              @click="addPlayer(player)"
            >
              <template #prepend>
                <v-avatar size="32" class="mr-2">
                  <v-img v-if="player.avatar" :src="player.avatar" />
                  <v-icon v-else>mdi-account</v-icon>
                </v-avatar>
              </template>
              
              <v-list-item-title>
                {{ player.alias }}
                <v-chip v-if="player.is_virtual" size="x-small" class="ml-1">
                  {{ $t('game.virtual') }}
                </v-chip>
              </v-list-item-title>
              
              <template #append>
                <v-btn
                  icon="mdi-plus"
                  size="small"
                  variant="text"
                  color="primary"
                />
              </template>
            </v-list-item>
            
            <v-list-item v-if="filteredAvailablePlayers.length === 0">
              <v-list-item-title class="text-grey">
                {{ $t('game.noPlayersFound') }}
              </v-list-item-title>
            </v-list-item>
          </v-list>
        </v-card-text>
      </v-card>
    </v-col>
    
    <!-- Right panel: Selected players -->
    <v-col cols="12" md="6">
      <v-card>
        <v-card-title>
          {{ $t('game.selectedPlayers') }}
          <v-chip 
            size="small" 
            class="ml-2"
            :color="playerCountColor"
          >
            {{ selectedPlayers.length }} / {{ maxPlayers }}
          </v-chip>
        </v-card-title>
        
        <v-card-text>
          <v-alert
            v-if="selectedPlayers.length < minPlayers"
            type="warning"
            density="compact"
            class="mb-2"
          >
            {{ $t('game.minPlayersWarning', { min: minPlayers }) }}
          </v-alert>
          
          <v-list density="compact" class="selected-players-list">
            <v-list-item
              v-for="(player, index) in selectedPlayers"
              :key="player.membership_id"
              :class="{ 
                'selected-moderator': player.membership_id === selectedModeratorId,
                'virtual-player': player.is_virtual 
              }"
              @click="toggleModerator(player)"
            >
              <template #prepend>
                <v-avatar size="32" class="mr-2">
                  <v-img v-if="player.avatar" :src="player.avatar" />
                  <v-icon v-else>mdi-account</v-icon>
                </v-avatar>
              </template>
              
              <v-list-item-title>
                {{ player.alias }}
                <v-chip 
                  v-if="player.membership_id === currentPlayerMembershipId" 
                  size="x-small" 
                  color="primary"
                  class="ml-1"
                >
                  {{ $t('game.you') }}
                </v-chip>
                <v-chip v-if="player.is_virtual" size="x-small" class="ml-1">
                  {{ $t('game.virtual') }}
                </v-chip>
                <v-icon 
                  v-if="hasModerator && player.membership_id === selectedModeratorId" 
                  size="small" 
                  color="warning"
                  class="ml-1"
                >
                  mdi-crown
                </v-icon>
              </v-list-item-title>
              
              <template #append>
                <v-btn
                  icon="mdi-close"
                  size="small"
                  variant="text"
                  color="error"
                  @click.stop="removePlayer(index)"
                />
              </template>
            </v-list-item>
            
            <v-list-item v-if="selectedPlayers.length === 0">
              <v-list-item-title class="text-grey">
                {{ $t('game.noPlayersSelected') }}
              </v-list-item-title>
            </v-list-item>
          </v-list>
          
          <div v-if="hasModerator" class="mt-2 text-caption text-grey">
            <v-icon size="small">mdi-information</v-icon>
            {{ $t('game.clickToSelectModerator') }}
          </div>
        </v-card-text>
      </v-card>
    </v-col>
  </v-row>
  
  <!-- Create virtual player dialog -->
  <CreateVirtualPlayerDialog
    v-model="showCreateDialog"
    :league-code="leagueCode"
    @created="onVirtualPlayerCreated"
  />
</template>

<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
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

// Player count color based on min/max
const playerCountColor = computed(() => {
  const count = selectedPlayers.value.length;
  if (count < props.minPlayers) return 'warning';
  if (count > props.maxPlayers) return 'error';
  return 'success';
});

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

<style scoped>
.available-players-list,
.selected-players-list {
  max-height: 400px;
  overflow-y: auto;
}

.virtual-player {
  opacity: 0.8;
}

.selected-moderator {
  background-color: rgba(var(--v-theme-warning), 0.1);
  border-left: 3px solid rgb(var(--v-theme-warning));
}
</style>
