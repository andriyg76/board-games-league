<template>
  <v-container>
    <v-row>
      <v-col>
        <h2>{{ isEditing ? $t('gameRounds.title') : $t('home.newGameRound') }}</h2>

        <div v-if="loadingRound" class="text-center pa-4">
          <v-progress-circular indeterminate color="primary" size="64" />
          <p class="mt-4">{{ $t('common.loading') }}</p>
        </div>

        <template v-else>
          <!-- Wizard Stepper -->
          <v-stepper v-if="!isEditing" v-model="step" :items="stepItems" class="mb-4">
            <template #[`item.1`]>
              <!-- Step 1: Select Game Type -->
              <v-card flat>
                <v-card-title>{{ $t('game.selectGameType') }}</v-card-title>
                <v-card-text>
                  <v-list>
                    <v-list-item
                      v-for="gameType in gameTypes"
                      :key="gameType.code"
                      :class="{ 'bg-primary-lighten-4': round.game_type === gameType.code }"
                      @click="selectGameType(gameType)"
                    >
                      <template #prepend>
                        <v-icon>{{ gameType.icon || 'mdi-dice-multiple' }}</v-icon>
                      </template>
                      <v-list-item-title>
                        {{ getLocalizedName(gameType.names) }}
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        {{ gameType.min_players }}-{{ gameType.max_players }} {{ $t('gameTypes.players') }}
                      </v-list-item-subtitle>
                      <template #append>
                        <v-icon v-if="round.game_type === gameType.code" color="primary">
                          mdi-check
                        </v-icon>
                      </template>
                    </v-list-item>
                  </v-list>
                </v-card-text>
                <v-card-actions>
                  <v-spacer />
                  <v-btn
                    color="primary"
                    :disabled="!round.game_type"
                    @click="goToStep2"
                  >
                    {{ $t('game.next') }}
                    <v-icon end>mdi-arrow-right</v-icon>
                  </v-btn>
                </v-card-actions>
              </v-card>
            </template>

            <template #[`item.2`]>
              <!-- Step 2: Select Players -->
              <v-card flat>
                <v-card-title>{{ $t('game.selectPlayers') }}</v-card-title>
                <v-card-text>
                  <div v-if="loadingSuggested" class="text-center pa-4">
                    <v-progress-circular indeterminate color="primary" />
                    <p class="mt-2">{{ $t('common.loading') }}</p>
                  </div>
                  <PlayerSelector
                    v-else
                    ref="playerSelectorRef"
                    :suggested-players="suggestedPlayers"
                    :min-players="selectedGameType?.min_players || 2"
                    :max-players="selectedGameType?.max_players || 10"
                    :has-moderator="hasModerator"
                    :league-code="leagueCode"
                    @update:selectedPlayers="onPlayersSelected"
                    @update:moderatorId="onModeratorSelected"
                  />
                </v-card-text>
                <v-card-actions>
                  <v-btn variant="text" @click="step = 1">
                    <v-icon start>mdi-arrow-left</v-icon>
                    {{ $t('game.back') }}
                  </v-btn>
                  <v-spacer />
                  <v-btn
                    color="primary"
                    :disabled="selectedPlayersList.length < (selectedGameType?.min_players || 2)"
                    @click="goToStep3"
                  >
                    {{ $t('game.next') }}
                    <v-icon end>mdi-arrow-right</v-icon>
                  </v-btn>
                </v-card-actions>
              </v-card>
            </template>

            <template #[`item.3`]>
              <!-- Step 3: Configure Round -->
              <v-card flat>
                <v-card-title>{{ $t('game.configureRound') }}</v-card-title>
                <v-card-text>
                  <v-text-field
                    v-model="round.name"
                    :label="$t('game.roundName')"
                    variant="outlined"
                    class="mb-4"
                  />

                  <!-- Wizard-specific options -->
                  <template v-if="isWizardGame">
                    <v-select
                      v-model="bidRestriction"
                      :items="bidRestrictions"
                      :item-title="(item) => $t(item.title)"
                      item-value="value"
                      :label="$t('wizard.bidRestriction')"
                      variant="outlined"
                      class="mb-4"
                    >
                      <template #prepend>
                        <v-icon>mdi-shield-alert</v-icon>
                      </template>
                    </v-select>

                    <h4 class="mb-2">{{ $t('wizard.selectFirstDealer') }}</h4>
                    <v-list density="compact" class="mb-4">
                      <v-list-item
                        v-for="(player, index) in roundPlayers"
                        :key="player.membership_id"
                        :class="{ 'bg-primary-lighten-4': firstDealerIndex === index }"
                        @click="firstDealerIndex = index"
                      >
                        <template #prepend>
                          <v-icon v-if="firstDealerIndex === index" color="primary">
                            mdi-cards-playing
                          </v-icon>
                          <span v-else class="text-grey ml-2 mr-3">{{ index + 1 }}</span>
                        </template>
                        <v-list-item-title>
                          {{ getPlayerAlias(player.membership_id) }}
                          <v-chip v-if="isCurrentPlayer(player.membership_id)" size="x-small" color="primary" class="ml-1">
                            {{ $t('game.you') }}
                          </v-chip>
                        </v-list-item-title>
                        <template #append v-if="firstDealerIndex === index">
                          <v-chip size="small" color="primary">
                            {{ $t('wizard.firstDealer') }}
                          </v-chip>
                        </template>
                      </v-list-item>
                    </v-list>

                    <!-- Wizard game summary -->
                    <v-card variant="outlined" class="pa-4">
                      <h4 class="text-subtitle-1 mb-2">{{ $t('wizard.gameSummary') }}</h4>
                      <v-row dense>
                        <v-col cols="6">
                          <div class="text-caption text-grey">{{ $t('game.selectedPlayers') }}</div>
                          <div class="text-body-1">{{ roundPlayers.length }}</div>
                        </v-col>
                        <v-col cols="6">
                          <div class="text-caption text-grey">{{ $t('wizard.rounds') }}</div>
                          <div class="text-body-1">{{ wizardMaxRounds }}</div>
                        </v-col>
                        <v-col cols="6">
                          <div class="text-caption text-grey">{{ $t('wizard.firstDealer') }}</div>
                          <div class="text-body-1">{{ getPlayerAlias(roundPlayers[firstDealerIndex]?.membership_id) || '-' }}</div>
                        </v-col>
                        <v-col cols="6">
                          <div class="text-caption text-grey">{{ $t('wizard.bidRestriction') }}</div>
                          <div class="text-body-1 text-truncate">{{ $t(bidRestrictions.find(r => r.value === bidRestriction)?.title || '') }}</div>
                        </v-col>
                      </v-row>
                    </v-card>
                  </template>

                  <!-- Non-wizard games: roles assignment -->
                  <template v-else-if="gameRoles.length > 0">
                    <h4 class="mb-2">{{ $t('game.assignRoles') }}</h4>
                    <v-table>
                      <thead>
                        <tr>
                          <th>{{ $t('leagues.player') }}</th>
                          <th>{{ $t('gameTypes.roles') }}</th>
                          <th v-if="hasModerator">{{ $t('roleTypes.moderator') }}</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(player, index) in roundPlayers" :key="player.membership_id">
                          <td>
                            {{ getPlayerAlias(player.membership_id) }}
                            <v-chip v-if="isCurrentPlayer(player.membership_id)" size="x-small" color="primary" class="ml-1">
                              {{ $t('game.you') }}
                            </v-chip>
                          </td>
                          <td>
                            <v-select
                              v-if="assignableRoles.length > 0"
                              v-model="player.role_key"
                              :items="assignableRoles"
                              item-title="name"
                              item-value="key"
                              density="compact"
                              variant="outlined"
                              hide-details
                            />
                            <span v-else class="text-grey">-</span>
                          </td>
                          <td v-if="hasModerator">
                            <v-checkbox
                              v-model="player.is_moderator"
                              hide-details
                              density="compact"
                              @update:modelValue="onModeratorCheckboxChange(index, $event)"
                            />
                          </td>
                        </tr>
                      </tbody>
                    </v-table>
                  </template>
                </v-card-text>
                <v-card-actions>
                  <v-btn variant="text" @click="step = 2">
                    <v-icon start>mdi-arrow-left</v-icon>
                    {{ $t('game.back') }}
                  </v-btn>
                  <v-spacer />
                  <v-btn
                    color="success"
                    :loading="saving"
                    @click="saveRound"
                  >
                    <v-icon start>mdi-check</v-icon>
                    {{ $t('game.startGame') }}
                  </v-btn>
                </v-card-actions>
              </v-card>
            </template>
          </v-stepper>

          <!-- Edit mode (simple form) -->
          <v-form v-else @submit.prevent="saveRound">
            <v-text-field
              v-model="round.name"
              :label="$t('game.roundName')"
              variant="outlined"
              class="mb-4"
            />

            <v-list v-if="round.players.length > 0">
              <v-subheader>{{ $t('game.selectedPlayers') }}</v-subheader>
              <v-list-item v-for="(player, index) in round.players" :key="index">
                <v-row>
                  <v-col cols="6">
                    <v-select
                      v-model="player.user_id"
                      :items="players"
                      item-title="alias"
                      item-value="code"
                      :label="$t('leagues.player')"
                      required
                    />
                  </v-col>
                  <v-col cols="3">
                    <v-checkbox
                      v-model="player.is_moderator"
                      :label="$t('roleTypes.moderator')"
                    />
                  </v-col>
                </v-row>
              </v-list-item>
            </v-list>

            <v-btn
              type="submit"
              color="success"
              class="mt-4"
              :loading="saving"
            >
              {{ $t('common.save') }}
            </v-btn>
          </v-form>
        </template>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
import { useGameStore } from '@/store/game';
import { usePlayerStore } from '@/store/player';
import { useLeagueStore } from '@/store/league';
import { useWizardStore } from '@/store/wizard';
import { useRouter, useRoute } from 'vue-router';
import { GameRound, GameRoundPlayer, Player, GameType, getLocalizedName, Role } from '@/api/GameApi';
import GameApi from '@/api/GameApi';
import LeagueApi, { SuggestedPlayer, SuggestedPlayersResponse } from '@/api/LeagueApi';
import PlayerSelector from '@/components/game/PlayerSelector.vue';
import { useI18n } from 'vue-i18n';
import { BidRestriction, GameVariant } from '@/wizard/types';

const props = defineProps<{
  id?: string;
}>();

const router = useRouter();
const route = useRoute();
const { locale } = useI18n();
const gameStore = useGameStore();
const playerStore = usePlayerStore();
const leagueStore = useLeagueStore();
const wizardStore = useWizardStore();

const loadingRound = ref(false);
const loadingSuggested = ref(false);
const saving = ref(false);
const step = ref(1);
const playerSelectorRef = ref<InstanceType<typeof PlayerSelector> | null>(null);

// Round data
const round = ref<GameRound>({
  code: '',
  name: '',
  game_type: '',
  start_time: new Date().toISOString(),
  players: [],
  version: 0
});

// Suggested players for player selection
const suggestedPlayers = ref<SuggestedPlayersResponse | null>(null);

// Selected players from PlayerSelector
const selectedPlayersList = ref<SuggestedPlayer[]>([]);
const selectedModeratorId = ref<string | null>(null);

// Round players for step 3
interface RoundPlayerSetup {
  membership_id: string;
  is_moderator: boolean;
  role_key?: string;
}
const roundPlayers = ref<RoundPlayerSetup[]>([]);

// Wizard-specific options
const bidRestriction = ref<BidRestriction>(BidRestriction.NO_RESTRICTIONS);
const firstDealerIndex = ref<number>(0);
const bidRestrictions = [
  { value: BidRestriction.NO_RESTRICTIONS, title: 'wizard.noRestrictions' },
  { value: BidRestriction.CANNOT_MATCH_CARDS, title: 'wizard.cannotMatchCards' },
  { value: BidRestriction.MUST_MATCH_CARDS, title: 'wizard.mustMatchCards' },
];

// Wizard step items
const stepItems = computed(() => [
  { title: 'game.selectGameType', value: 1 },
  { title: 'game.selectPlayers', value: 2 },
  { title: 'game.configureRound', value: 3 },
]);

// Get league code from route or store
const leagueCode = computed(() => {
  return (route.query.league as string) || leagueStore.currentLeague?.code || '';
});

// Old players list for edit mode
const players = computed(() => {
  const playerList = playerStore.players || [];
  return playerList.map((p: Player) => ({
    code: p.code,
    alias: p.alias,
    title: p.alias,
    props: {
      avatar: p.avatar,
      prependAvatar: p.avatar,
    }
  }));
});

const gameTypes = computed(() => gameStore.gameTypes);
const isEditing = computed(() => !!round.value.code);

const selectedGameType = computed(() =>
  gameTypes.value.find(gt => gt.code === round.value.game_type)
);

// Check if selected game type is Wizard
const isWizardGame = computed(() => {
  if (!selectedGameType.value) return false;
  // Check by key 'wizard' (from games.yaml)
  return selectedGameType.value.code === 'wizard' || 
         (selectedGameType.value as unknown as { key?: string }).key === 'wizard';
});

// Max rounds for Wizard game (60 cards / players)
const wizardMaxRounds = computed(() => {
  if (selectedPlayersList.value.length === 0) return 0;
  return Math.floor(60 / selectedPlayersList.value.length);
});

// Check if game type has moderator role
const hasModerator = computed(() => {
  if (!selectedGameType.value?.roles) return false;
  return selectedGameType.value.roles.some((r: Role) => r.role_type === 'moderator');
});

// Get assignable roles (excluding moderator - handled separately)
const assignableRoles = computed(() => {
  if (!selectedGameType.value?.roles) return [];
  return selectedGameType.value.roles
    .filter((r: Role) => r.role_type !== 'moderator')
    .map((r: Role) => ({
      key: r.key,
      name: getLocalizedName(r.names, locale.value),
    }));
});

// Get all roles for the game
const gameRoles = computed(() => {
  if (!selectedGameType.value?.roles) return [];
  return selectedGameType.value.roles;
});

// Select a game type
const selectGameType = (gameType: GameType) => {
  round.value.game_type = gameType.code;
};

// Go to step 2 - load suggested players
const goToStep2 = async () => {
  if (!round.value.game_type) return;
  
  step.value = 2;
  
  if (leagueCode.value && !suggestedPlayers.value) {
    loadingSuggested.value = true;
    try {
      suggestedPlayers.value = await LeagueApi.getSuggestedPlayers(leagueCode.value);
    } catch (error) {
      console.error('Failed to load suggested players:', error);
    } finally {
      loadingSuggested.value = false;
    }
  }
};

// Go to step 3 - prepare round players
const goToStep3 = () => {
  // Convert selected players to round players
  roundPlayers.value = selectedPlayersList.value.map(player => ({
    membership_id: player.membership_id,
    is_moderator: hasModerator.value && player.membership_id === selectedModeratorId.value,
    role_key: undefined,
  }));
  
  step.value = 3;
};

// Handle players selection update
const onPlayersSelected = (players: SuggestedPlayer[]) => {
  selectedPlayersList.value = players;
};

// Handle moderator selection update
const onModeratorSelected = (moderatorId: string | null) => {
  selectedModeratorId.value = moderatorId;
};

// Handle moderator checkbox change (only one can be selected)
const onModeratorCheckboxChange = (index: number, isChecked: boolean | null) => {
  if (isChecked) {
    // Uncheck all others
    roundPlayers.value.forEach((p, i) => {
      if (i !== index) p.is_moderator = false;
    });
  }
};

// Get player alias by membership ID
const getPlayerAlias = (membershipId: string): string => {
  const player = selectedPlayersList.value.find(p => p.membership_id === membershipId);
  return player?.alias || membershipId;
};

// Check if player is current user
const isCurrentPlayer = (membershipId: string): boolean => {
  return suggestedPlayers.value?.current_player?.membership_id === membershipId;
};

// Load game round for editing
const loadGameRound = async () => {
  if (!props.id) return;

  loadingRound.value = true;
  try {
    const loadedRound = await GameApi.getGameRound(props.id);
    round.value = {
      code: props.id,
      name: loadedRound.name,
      game_type: loadedRound.game_type,
      start_time: loadedRound.start_time,
      players: loadedRound.players,
      version: loadedRound.version
    };
  } catch (error) {
    console.error('Error loading game round:', error);
  } finally {
    loadingRound.value = false;
  }
};

// Save round
const saveRound = async () => {
  saving.value = true;
  try {
    if (isEditing.value) {
      const savedRound = await gameStore.updateRound(round.value);
      await router.push({
        name: 'GameRounds',
        params: { code: savedRound.code }
      });
    } else if (isWizardGame.value) {
      // Create wizard game
      const wizardRequest = {
        league_id: leagueCode.value,
        game_name: round.value.name || `Wizard ${new Date().toLocaleDateString()}`,
        bid_restriction: bidRestriction.value,
        game_variant: GameVariant.STANDARD,
        first_dealer_index: firstDealerIndex.value,
        player_membership_ids: roundPlayers.value.map(p => p.membership_id),
      };

      await wizardStore.createGame(wizardRequest);
      
      // Navigate to wizard game
      if (wizardStore.currentGame) {
        await router.push(`/ui/wizard/${wizardStore.currentGame.code}`);
      } else {
        await router.push({ name: 'GameRounds' });
      }
    } else {
      // Create new generic game round with players
      const players: GameRoundPlayer[] = roundPlayers.value.map((p, index) => ({
        membership_id: p.membership_id,
        user_id: '',  // Will be resolved on backend
        score: 0,
        is_moderator: p.is_moderator,
        team_name: p.role_key,  // Using team_name for role for now
        position: index + 1,
      }));

      const newRound: GameRound = {
        name: round.value.name || `Game ${new Date().toLocaleDateString()}`,
        game_type: round.value.game_type,
        start_time: new Date().toISOString(),
        players,
        version: 0,
      };

      const savedRound = await gameStore.addActiveRound(newRound);
      await router.push({
        name: 'GameRounds',
        params: { code: savedRound.code }
      });
    }
  } catch (error) {
    console.error('Error saving game round:', error);
  } finally {
    saving.value = false;
  }
};

onMounted(async () => {
  try {
    // Load game types first
    if (gameStore.gameTypes.length === 0) {
      await gameStore.loadGameTypes();
    }

    // Load players for edit mode
    if (!playerStore.players || playerStore.players.length === 0) {
      await GameApi.listPlayers().then(p => {
        playerStore.players = p;
      });
    }

    // If editing, load the game round
    if (props.id) {
      await loadGameRound();
    } else {
      // Check for preselected game type from query params
      const preselectedGameType = route.query.gameType as string;
      if (preselectedGameType) {
        const gameType = gameTypes.value.find(
          gt => gt.code === preselectedGameType || 
                (gt as unknown as { key?: string }).key === preselectedGameType
        );
        if (gameType) {
          round.value.game_type = gameType.code;
          // Automatically go to step 2 if game type is preselected
          await goToStep2();
        }
      }
    }
  } catch (error) {
    console.error('Failed to load data:', error);
  }
});
</script>

<style scoped>
.bg-primary-lighten-4 {
  background-color: rgba(var(--v-theme-primary), 0.1);
}
</style>
