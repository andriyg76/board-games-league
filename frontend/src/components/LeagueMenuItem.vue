<template>
  <n-dropdown v-if="shouldShowMenu" :options="menuOptions" trigger="hover" @select="handleSelect">
    <n-button quaternary>
      {{ currentLeagueName || t('leagues.menu') }}
      <template #icon>
        <n-icon><ChevronDownIcon /></n-icon>
      </template>
    </n-button>
  </n-dropdown>
</template>

<script lang="ts" setup>
import { h, computed, onMounted } from 'vue';
import { NButton, NDropdown, NIcon } from 'naive-ui';
import { ChevronDown as ChevronDownIcon, Checkmark as CheckmarkIcon } from '@vicons/ionicons5';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { useLeagueStore } from '@/store/league';
import { useUserStore } from '@/store/user';

const { t } = useI18n();
const router = useRouter();
const leagueStore = useLeagueStore();
const userStore = useUserStore();

const currentLeagueName = computed(() => {
  return leagueStore.currentLeague?.name || null;
});

const shouldShowMenu = computed(() => {
  const activeLeagues = leagueStore.activeLeagues;
  // Always show menu for superadmin, otherwise hide if there are 0 or 1 leagues
  return userStore.isSuperAdmin || activeLeagues.length > 1;
});

const menuOptions = computed(() => {
  // For superadmin, show all leagues (active + archived), otherwise only active
  const leagues = userStore.isSuperAdmin 
    ? leagueStore.leagues 
    : leagueStore.activeLeagues;
  const currentCode = leagueStore.currentLeagueCode;
  
  if (leagues.length === 0) {
    return [
      {
        label: t('leagues.noActiveLeagues'),
        key: 'no-leagues',
        disabled: true,
      },
    ];
  }

  // Sort leagues: current league first, then others
  const sortedLeagues = [...leagues].sort((a, b) => {
    if (a.code === currentCode) return -1;
    if (b.code === currentCode) return 1;
    return 0;
  });

  const leagueOptions = sortedLeagues.map(league => ({
    label: () => h('div', { style: 'display: flex; align-items: center; gap: 8px;' }, [
      league.code === currentCode 
        ? h(NIcon, { size: 16, color: '#18a058' }, { default: () => h(CheckmarkIcon) })
        : h('span', { style: 'width: 16px;' }),
      h('span', league.name),
      league.status === 'archived' 
        ? h('span', { style: 'font-size: 0.75rem; opacity: 0.6; margin-left: 4px;' }, `(${t('leagues.archived')})`)
        : null,
    ]),
    key: league.code,
  }));

  // Add divider and "All Leagues" link at the end
  return [
    ...leagueOptions,
    {
      type: 'divider',
      key: 'divider',
    },
    {
      label: t('nav.leagues'),
      key: 'all-leagues',
    },
  ];
});

const handleSelect = (key: string) => {
  if (key === 'no-leagues') return;
  if (key === 'all-leagues') {
    router.push({ name: 'Leagues' });
    return;
  }
  router.push({ name: 'LeagueDetails', params: { code: key } });
};

// Load leagues on mount
onMounted(async () => {
  if (leagueStore.leagues.length === 0) {
    try {
      await leagueStore.loadLeagues();
    } catch (error) {
      console.error('Error loading leagues:', error);
    }
  }
});
</script>

