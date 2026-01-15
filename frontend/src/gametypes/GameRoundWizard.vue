<template>
  <n-grid :cols="24" :x-gap="16">
    <n-gi :span="24">
      <h2 style="font-size: 2rem; margin-bottom: 16px;">{{ $t('home.newGameRound') }}</h2>

      <n-spin v-if="loading" size="large" style="display: flex; justify-content: center; padding: 64px;">
        <template #description>
          {{ $t('common.loading') }}
        </template>
      </n-spin>

      <div v-else>
        <!-- Step Indicator -->
        <n-steps :current="step - 1" :status="'process'" style="margin-bottom: 24px;">
          <n-step
            v-for="(item, index) in stepItems"
            :key="item.value"
            :title="$t(item.title)"
          />
        </n-steps>

        <!-- Step 1: Select Game Type -->
        <div v-if="step === 1">
          <Step1GameType
            :game-types="gameTypes"
            :selected-game-type="selectedGameTypeCode"
            @select="selectGameType"
            @next="goToStep2"
          />
        </div>

        <!-- Step 2: Select Players -->
        <div v-if="step === 2">
          <Step2Players
            :suggested-players="suggestedPlayers"
            :min-players="selectedGameType?.min_players || 2"
            :max-players="selectedGameType?.max_players || 10"
            :has-moderator="hasModerator"
            :league-code="leagueCode"
            :loading="loadingSuggested"
            :saving="saving"
            :can-proceed="canProceedFromStep2"
            @update:selected-players="onPlayersSelected"
            @update:moderator-id="onModeratorSelected"
            @back="step = 1"
            @next="goToStep3"
          />
        </div>

        <!-- Step 3: Configure Round / Wizard Config -->
        <div v-if="step === 3">
          <WizardGameConfig
            v-if="isWizardGame"
            v-model:game-name="roundName"
            v-model:bid-restriction="bidRestriction"
            v-model:first-dealer-index="firstDealerIndex"
            :players="wizardPlayers"
            :saving="saving"
            @back="step = 2"
            @start="startWizardGame"
          />
          <Step3Roles
            v-else
            v-model:round-name="roundName"
            :players="rolePlayers"
            :roles="gameRoles"
            :assignable-roles="assignableRoles"
            :has-moderator="hasModerator"
            :saving="saving"
            @update-role="updatePlayerRole"
            @update-moderator="updatePlayerModerator"
            @back="step = 2"
            @save="saveRoles"
            @next="goToStep4"
          />
        </div>

        <!-- Step 4: Enter Scores (non-wizard only) -->
        <div v-if="step === 4 && !isWizardGame">
          <Step4Scoring
            :players="scoringPlayers"
            :scores="playerScores"
            :positions="playerPositions"
            :saving="saving"
            @update-score="updatePlayerScore"
            @update-position="updatePlayerPosition"
            @back="step = 3"
            @save="saveScores"
            @finish="finishGame"
          />
        </div>
      </div>
    </n-gi>
  </n-grid>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, watch } from 'vue';
import { NGrid, NGi, NSpin, NSteps, NStep, useDialog, useMessage } from 'naive-ui';
import { useRouter, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useGameStore } from '@/store/game';
import { useLeagueStore } from '@/store/league';
import { useWizardStore } from '@/store/wizard';
import { useUserStore } from '@/store/user';
import GameApi, { GameType, Role, getLocalizedName } from '@/api/GameApi';
import LeagueApi, { SuggestedPlayer, SuggestedPlayersResponse } from '@/api/LeagueApi';
import { BidRestriction, GameVariant } from '@/wizard/types';
import { useErrorHandler } from '@/composables/useErrorHandler';

import Step1GameType from './steps/Step1GameType.vue';
import Step2Players from './steps/Step2Players.vue';
import Step3Roles from './steps/Step3Roles.vue';
import Step4Scoring from './steps/Step4Scoring.vue';
import WizardGameConfig from '@/wizard/WizardGameConfig.vue';

const props = defineProps<{
  id?: string;
}>();

const router = useRouter();
const route = useRoute();
const { locale, t } = useI18n();
const gameStore = useGameStore();
const leagueStore = useLeagueStore();
const wizardStore = useWizardStore();
const userStore = useUserStore();
const dialog = useDialog();
const message = useMessage();
const { handleError, showSuccess } = useErrorHandler();

// State
const loading = ref(false);
const loadingSuggested = ref(false);
const saving = ref(false);
const step = ref(1);

// Round data
const roundCode = ref<string>('');
const roundName = ref<string>('');
const roundVersion = ref<number>(0);
const selectedGameTypeCode = ref<string>('');

// Player selection
const suggestedPlayers = ref<SuggestedPlayersResponse | null>(null);
const selectedPlayersList = ref<SuggestedPlayer[]>([]);
const selectedModeratorId = ref<string | null>(null);

// Round players for steps 3-4
interface RoundPlayer {
  membership_code: string;
  alias: string;
  is_moderator: boolean;
  role_key?: string;
  isCurrentUser?: boolean;
}
const roundPlayers = ref<RoundPlayer[]>([]);

// Wizard-specific
const bidRestriction = ref<BidRestriction>(BidRestriction.NO_RESTRICTIONS);
const firstDealerIndex = ref<number>(0);

// Scoring (step 4)
const playerScores = ref<Record<string, number>>({});
const playerPositions = ref<Record<string, number>>({});

// Computed
const leagueCode = computed(() => {
  return (route.query.league as string) || leagueStore.currentLeagueCode || '';
});

const gameTypes = computed(() => gameStore.gameTypes);

const selectedGameType = computed(() =>
  gameTypes.value.find(gt => gt.code === selectedGameTypeCode.value)
);

const isWizardGame = computed(() => {
  if (!selectedGameType.value) return false;
  return selectedGameType.value.code === 'wizard' ||
    (selectedGameType.value as unknown as { key?: string }).key === 'wizard';
});

const hasModerator = computed(() => {
  if (!selectedGameType.value?.roles) return false;
  return selectedGameType.value.roles.some((r: Role) => r.role_type === 'moderator');
});

const gameRoles = computed(() => {
  if (!selectedGameType.value?.roles) return [];
  return selectedGameType.value.roles;
});

const assignableRoles = computed(() => {
  if (!selectedGameType.value?.roles) return [];
  return selectedGameType.value.roles
    .filter((r: Role) => r.role_type !== 'moderator')
    .map((r: Role) => ({
      key: r.key,
      name: getLocalizedName(r.names, locale.value),
    }));
});

const stepItems = computed(() => {
  const items = [
    { title: 'game.selectGameType', value: 1 },
    { title: 'game.selectPlayers', value: 2 },
    { title: 'game.configureRound', value: 3 },
  ];
  if (!isWizardGame.value) {
    items.push({ title: 'game.enterScores', value: 4 });
  }
  return items;
});

const canProceedFromStep2 = computed(() => {
  return selectedPlayersList.value.length >= (selectedGameType.value?.min_players || 2);
});

// Players formatted for different steps
const wizardPlayers = computed(() => {
  return roundPlayers.value.map(p => ({
    membership_id: p.membership_code, // WizardPlayer expects membership_id
    membership_code: p.membership_code,
    alias: p.alias,
    isCurrentUser: p.isCurrentUser,
  }));
});

const rolePlayers = computed(() => roundPlayers.value);

const scoringPlayers = computed(() => {
  return roundPlayers.value.map(p => ({
    membership_code: p.membership_code,
    alias: p.alias,
    is_moderator: p.is_moderator,
  }));
});

// Methods
const selectGameType = (gameType: GameType) => {
  selectedGameTypeCode.value = gameType.code;
};

const goToStep2 = async () => {
  if (!selectedGameTypeCode.value) return;

  step.value = 2;

  if (leagueCode.value && !suggestedPlayers.value) {
    loadingSuggested.value = true;
    try {
      suggestedPlayers.value = await LeagueApi.getSuggestedPlayers(leagueCode.value);
      
      // Check if superadmin can create membership
      if (suggestedPlayers.value?.can_create_membership && userStore.isSuperAdmin) {
        showCreateMembershipDialog();
      }
    } catch (error) {
      handleError(error, t('errors.loadingData'));
    } finally {
      loadingSuggested.value = false;
    }
  }
};

const showCreateMembershipDialog = () => {
  dialog.warning({
    title: t('leagues.createMembershipTitle'),
    content: t('leagues.createMembershipMessage'),
    positiveText: t('common.create'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      await createMembership();
    },
  });
};

const createMembership = async () => {
  if (!leagueCode.value) return;
  
  try {
    await LeagueApi.createMembership(leagueCode.value);
    message.success(t('leagues.membershipCreated'));
    
    // Reload suggested players to get current_player
    loadingSuggested.value = true;
    try {
      suggestedPlayers.value = await LeagueApi.getSuggestedPlayers(leagueCode.value);
    } catch (error) {
      handleError(error, t('errors.loadingData'));
    } finally {
      loadingSuggested.value = false;
    }
  } catch (error) {
    handleError(error, t('errors.savingData'));
  }
};

const onPlayersSelected = (players: SuggestedPlayer[]) => {
  selectedPlayersList.value = players;
};

const onModeratorSelected = (moderatorId: string | null) => {
  selectedModeratorId.value = moderatorId;
};

// Helper to get membership code (support both membership_code and legacy membership_id)
const getMembershipCode = (player: SuggestedPlayer): string => {
  return player.membership_code || player.membership_id || '';
};

const goToStep3 = async () => {
  // Build round players
  const currentPlayerCode = suggestedPlayers.value?.current_player 
    ? getMembershipCode(suggestedPlayers.value.current_player)
    : '';
  
  roundPlayers.value = selectedPlayersList.value.map(player => {
    const code = getMembershipCode(player);
    return {
      membership_code: code,
      alias: player.alias,
      is_moderator: hasModerator.value && code === selectedModeratorId.value,
      role_key: undefined,
      isCurrentUser: code === currentPlayerCode,
    };
  });

  // For wizard games, just go to step 3
  if (isWizardGame.value) {
    step.value = 3;
    return;
  }

  // For other games, save round to server first
  if (!roundCode.value && leagueCode.value) {
    saving.value = true;
    try {
      const players = selectedPlayersList.value.map((player, index) => ({
        membership_code: getMembershipCode(player),
        position: index + 1,
        is_moderator: hasModerator.value && getMembershipCode(player) === selectedModeratorId.value,
      }));

      const savedRound = await GameApi.createLeagueGameRound(leagueCode.value, {
        name: roundName.value || `Game ${new Date().toLocaleDateString()}`,
        type: selectedGameTypeCode.value,
        players,
      });

      roundCode.value = savedRound.code || '';
      roundVersion.value = savedRound.version;

      // Update URL for bookmarking
      await router.replace({
        name: 'EditGameRound',
        params: { id: savedRound.code },
        query: route.query,
      });
    } catch (error) {
      handleError(error, t('errors.savingData'));
      saving.value = false;
      return;
    } finally {
      saving.value = false;
    }
  }

  step.value = 3;
};

const updatePlayerRole = (index: number, roleKey: string) => {
  if (roundPlayers.value[index]) {
    roundPlayers.value[index].role_key = roleKey;
  }
};

const updatePlayerModerator = (index: number, isModerator: boolean) => {
  if (isModerator) {
    // Uncheck all others
    roundPlayers.value.forEach((p, i) => {
      p.is_moderator = i === index;
    });
  } else {
    roundPlayers.value[index].is_moderator = false;
  }
};

const saveRoles = async () => {
  if (!roundCode.value || !leagueCode.value) return;

  saving.value = true;
  try {
    const players = roundPlayers.value.map(p => ({
      membership_code: p.membership_code,
      role_key: p.role_key,
      is_moderator: p.is_moderator,
    }));

    const updatedRound = await GameApi.updateRoles(leagueCode.value, roundCode.value, players);
    roundVersion.value = updatedRound.version;
  } catch (error) {
    handleError(error, t('errors.savingData'));
  } finally {
    saving.value = false;
  }
};

const goToStep4 = async () => {
  await saveRoles();

  // Initialize scores and positions
  roundPlayers.value.forEach((player, index) => {
    if (playerScores.value[player.membership_code] === undefined) {
      playerScores.value[player.membership_code] = 0;
    }
    if (playerPositions.value[player.membership_code] === undefined) {
      playerPositions.value[player.membership_code] = index + 1;
    }
  });

  step.value = 4;
};

const updatePlayerScore = (membershipId: string, score: number) => {
  playerScores.value[membershipId] = score;
};

const updatePlayerPosition = (membershipId: string, position: number) => {
  playerPositions.value[membershipId] = position;
};

const saveScores = async () => {
  if (!roundCode.value || !leagueCode.value) return;

  saving.value = true;
  try {
    const updatedRound = await GameApi.updateScores(leagueCode.value, roundCode.value, playerScores.value);
    roundVersion.value = updatedRound.version;
  } catch (error) {
    handleError(error, t('errors.savingData'));
  } finally {
    saving.value = false;
  }
};

const finishGame = async () => {
  if (!roundCode.value || !leagueCode.value) return;

  saving.value = true;
  try {
    await saveScores();
    await GameApi.finalizeGameRound(leagueCode.value, roundCode.value, {
      player_scores: playerScores.value,
    });
    showSuccess(t('game.gameFinished'));
    await router.push({ name: 'GameRounds' });
  } catch (error) {
    handleError(error, t('errors.savingData'));
  } finally {
    saving.value = false;
  }
};

const startWizardGame = async () => {
  saving.value = true;
  try {
    const wizardRequest = {
      game_name: roundName.value || `Wizard ${new Date().toLocaleDateString()}`,
      bid_restriction: bidRestriction.value,
      game_variant: GameVariant.STANDARD,
      first_dealer_index: firstDealerIndex.value,
      player_membership_codes: roundPlayers.value.map(p => p.membership_code),
    };

    await wizardStore.createGame(leagueCode.value, wizardRequest);

    if (wizardStore.currentGame) {
      showSuccess(t('wizard.gameCreated'));
      await router.push({
        name: 'WizardGame',
        params: { code: wizardStore.currentGame.code },
        query: { league: leagueCode.value }
      });
    } else {
      await router.push({ name: 'GameRounds' });
    }
  } catch (error) {
    handleError(error, t('errors.savingData'));
  } finally {
    saving.value = false;
  }
};

const loadExistingRound = async () => {
  if (!props.id) return;

  if (!leagueCode.value) {
    handleError(new Error('No league selected'), t('errors.loadingData'));
    return;
  }

  loading.value = true;
  try {
    const loadedRound = await GameApi.getGameRound(leagueCode.value, props.id);
    
    // Redirect to edit page for completed rounds
    if (loadedRound.status === 'completed') {
      await router.replace({ 
        name: 'EditCompletedGameRound', 
        params: { id: props.id } 
      });
      return;
    }

    roundCode.value = props.id;
    roundName.value = loadedRound.name;
    selectedGameTypeCode.value = loadedRound.game_type;
    roundVersion.value = loadedRound.version;

    // Build round players from loaded data
    roundPlayers.value = loadedRound.players.map(p => {
      // Support both membership_code (new) and membership_id (legacy)
      const membershipCode = (p as any).membership_code || (p as any).membership_id || '';
      return {
        membership_code: membershipCode,
        alias: membershipCode, // Will be resolved later
        is_moderator: p.is_moderator,
        role_key: p.label_name,
        isCurrentUser: false,
      };
    });

    // Set step based on status
    const status = loadedRound.status;
    if (status === 'players_selected') {
      step.value = 3;
    } else if (status === 'in_progress') {
      step.value = 3;
    } else if (status === 'scoring') {
      // Initialize scores
      loadedRound.players.forEach((p, index) => {
        // Support both membership_code (new) and membership_id (legacy)
        const code = (p as any).membership_code || (p as any).membership_id || '';
        playerScores.value[code] = p.score || 0;
        playerPositions.value[code] = p.position || index + 1;
      });
      step.value = 4;
    }
  } catch (error) {
    handleError(error, t('errors.loadingData'));
  } finally {
    loading.value = false;
  }
};

onMounted(async () => {
  try {
    if (gameStore.gameTypes.length === 0) {
      await gameStore.loadGameTypes();
    }

    if (props.id) {
      await loadExistingRound();
    } else {
      // Check for preselected game type
      const preselectedGameType = route.query.gameType as string;
      if (preselectedGameType) {
        const gameType = gameTypes.value.find(
          gt => gt.code === preselectedGameType ||
            (gt as unknown as { key?: string }).key === preselectedGameType
        );
        if (gameType) {
          selectedGameTypeCode.value = gameType.code;
          await goToStep2();
        }
      }
    }
  } catch (error) {
    handleError(error, t('errors.loadingData'));
  }
});
</script>
