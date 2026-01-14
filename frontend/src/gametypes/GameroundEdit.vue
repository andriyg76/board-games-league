<template>
  <n-grid :cols="24" :x-gap="16">
    <n-gi :span="24">
      <h2 style="font-size: 2rem; margin-bottom: 16px;">{{ $t('gameRounds.title') }}</h2>

      <n-spin v-if="loading" size="large" style="display: flex; justify-content: center; padding: 64px;">
        <template #description>
          {{ $t('common.loading') }}
        </template>
      </n-spin>

      <n-form v-else @submit.prevent="saveRound">
        <n-form-item :label="$t('game.roundName')">
          <n-input v-model:value="round.name" />
        </n-form-item>

        <n-card v-if="round.players.length > 0" style="margin-bottom: 16px;">
          <template #header>
            {{ $t('game.selectedPlayers') }}
          </template>
          <n-list>
            <n-list-item v-for="(player, index) in round.players" :key="index">
              <n-grid :cols="24" :x-gap="8" style="width: 100%;">
                <n-gi :span="24" :responsive="{ m: 6 }">
                  <span>{{ getPlayerAlias(player) }}</span>
                </n-gi>
                <n-gi :span="24" :responsive="{ m: 4 }">
                  <n-input-number
                    v-model:value="player.score"
                    :placeholder="$t('leagues.points')"
                    :min="0"
                    size="small"
                    style="width: 100%;"
                  />
                </n-gi>
                <n-gi :span="24" :responsive="{ m: 3 }">
                  <n-input-number
                    v-model:value="player.position"
                    :placeholder="$t('game.position')"
                    :min="1"
                    size="small"
                    style="width: 100%;"
                  />
                </n-gi>
                <n-gi :span="24" :responsive="{ m: 3 }">
                  <n-checkbox v-model:checked="player.is_moderator">
                    {{ $t('roleTypes.moderator') }}
                  </n-checkbox>
                </n-gi>
              </n-grid>
            </n-list-item>
          </n-list>
        </n-card>

        <div style="display: flex; justify-content: space-between; align-items: center; margin-top: 16px;">
          <n-button quaternary @click="goBack">
            {{ $t('common.cancel') }}
          </n-button>
          <n-button
            type="success"
            :loading="saving"
            @click="saveRound"
          >
            <template #icon>
              <n-icon><SaveIcon /></n-icon>
            </template>
            {{ $t('common.save') }}
          </n-button>
        </div>
      </n-form>
    </n-gi>
  </n-grid>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { NGrid, NGi, NSpin, NForm, NFormItem, NInput, NCard, NList, NListItem, NInputNumber, NCheckbox, NButton, NIcon } from 'naive-ui';
import { Save as SaveIcon } from '@vicons/ionicons5';
import { useRouter } from 'vue-router';
import { useGameStore } from '@/store/game';
import { usePlayerStore } from '@/store/player';
import GameApi, { GameRound, GameRoundPlayer } from '@/api/GameApi';

const props = defineProps<{
  id: string;
}>();

const router = useRouter();
const gameStore = useGameStore();
const playerStore = usePlayerStore();

const loading = ref(false);
const saving = ref(false);
const round = ref<GameRound>({
  code: '',
  name: '',
  game_type: '',
  start_time: new Date().toISOString(),
  players: [],
  version: 0
});

const getPlayerAlias = (player: GameRoundPlayer): string => {
  // Try to find player in player store
  const found = playerStore.players?.find(p => p.code === player.user_id);
  if (found) return found.alias;
  
  // Fallback to membership_id
  return player.membership_id || player.user_id || 'Unknown';
};

const loadRound = async () => {
  loading.value = true;
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
    loading.value = false;
  }
};

const saveRound = async () => {
  saving.value = true;
  try {
    await gameStore.updateRound(round.value);
    await router.push({ name: 'GameRounds' });
  } catch (error) {
    console.error('Error saving game round:', error);
  } finally {
    saving.value = false;
  }
};

const goBack = () => {
  router.push({ name: 'GameRounds' });
};

onMounted(async () => {
  // Load players for alias lookup
  if (!playerStore.players || playerStore.players.length === 0) {
    try {
      const players = await GameApi.listPlayers();
      playerStore.players = players;
    } catch (error) {
      console.error('Failed to load players:', error);
    }
  }

  await loadRound();
});
</script>

