<template>
  <div class="mobile-entry">
    <n-spin v-if="loading" size="large" />
    <div v-else class="mobile-entry__message">
      {{ message }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { NSpin } from 'naive-ui';
import { useLeagueStore } from '@/store/league';
import { useUserStore } from '@/store/user';

const router = useRouter();
const leagueStore = useLeagueStore();
const userStore = useUserStore();

const loading = ref(true);
const message = ref('Loading...');

const resolveEntry = async () => {
  if (!userStore.loggedIn) {
    await router.replace({ name: 'MobileLogin', query: { redirect: '/m' } });
    return;
  }

  if (leagueStore.leagues.length === 0) {
    await leagueStore.loadLeagues();
  }

  const activeLeagues = leagueStore.activeLeagues;
  const savedLeagueCode = leagueStore.getSavedLeagueCode();
  let targetLeagueCode: string | null = null;

  if (savedLeagueCode) {
    const savedLeague = activeLeagues.find((league) => league.code === savedLeagueCode);
    if (savedLeague) {
      targetLeagueCode = savedLeagueCode;
    }
  }

  if (!targetLeagueCode) {
    if (activeLeagues.length > 1) {
      await router.replace({ name: 'MobileLeagueSelect' });
      return;
    }

    if (activeLeagues.length === 1) {
      targetLeagueCode = activeLeagues[0].code;
    }
  }

  if (targetLeagueCode) {
    await leagueStore.setCurrentLeague(targetLeagueCode);
    await router.replace({ name: 'MobileLeagueHome' });
    return;
  }

  message.value = 'No leagues available yet.';
};

onMounted(async () => {
  try {
    await resolveEntry();
  } catch (error) {
    console.error('Failed to resolve mobile entry:', error);
    message.value = 'Unable to load leagues.';
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.mobile-entry {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  padding: 32px;
  text-align: center;
}

.mobile-entry__message {
  font-size: 1rem;
  opacity: 0.7;
}
</style>
