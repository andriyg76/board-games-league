<template>
  <div class="mobile-league-select">
    <h1 class="mobile-league-select__title">{{ t('leagues.title') }}</h1>
    <n-spin v-if="loading" size="large" />

    <n-alert v-else-if="activeLeagues.length === 0" type="info">
      {{ t('leagues.noActiveLeagues') }}
    </n-alert>

    <n-list v-else>
      <n-list-item
        v-for="league in activeLeagues"
        :key="league.code"
        clickable
        @click="selectLeague(league.code)"
      >
        <div class="mobile-league-select__item">
          <div class="mobile-league-select__name">{{ league.name }}</div>
        </div>
      </n-list-item>
    </n-list>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { NSpin, NList, NListItem, NAlert } from 'naive-ui';
import { useLeagueStore } from '@/store/league';
import { useUserStore } from '@/store/user';

const { t } = useI18n();
const router = useRouter();
const leagueStore = useLeagueStore();
const userStore = useUserStore();

const activeLeagues = computed(() => leagueStore.activeLeagues);
const loading = computed(() => leagueStore.isLoading);

const redirectToLogin = () => {
  router.replace({ name: 'MobileLogin', query: { redirect: '/m/league/select' } });
};

const loadLeagues = async () => {
  if (!userStore.loggedIn) {
    redirectToLogin();
    return;
  }
  if (leagueStore.leagues.length === 0) {
    await leagueStore.loadLeagues();
  }

  const savedLeagueCode = leagueStore.getSavedLeagueCode();
  if (savedLeagueCode) {
    const savedLeague = activeLeagues.value.find((league) => league.code === savedLeagueCode);
    if (savedLeague) {
      await leagueStore.setCurrentLeague(savedLeagueCode);
      await router.replace({ name: 'MobileLeagueHome' });
      return;
    }
  }

  if (activeLeagues.value.length === 1) {
    await leagueStore.setCurrentLeague(activeLeagues.value[0].code);
    await router.replace({ name: 'MobileLeagueHome' });
  }
};

const selectLeague = async (code: string) => {
  await leagueStore.setCurrentLeague(code);
  await router.push({ name: 'MobileLeagueHome' });
};

onMounted(async () => {
  try {
    await loadLeagues();
  } catch (error) {
    console.error('Failed to load leagues:', error);
  }
});
</script>

<style scoped>
.mobile-league-select {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 24px;
}

.mobile-league-select__title {
  font-size: 1.5rem;
  margin: 0;
}

.mobile-league-select__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.mobile-league-select__name {
  font-weight: 500;
}
</style>
