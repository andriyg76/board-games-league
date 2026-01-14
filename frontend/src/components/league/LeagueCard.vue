<template>
  <n-list-item
    clickable
    :style="{ backgroundColor: league.status === 'archived' ? 'rgba(0, 0, 0, 0.02)' : undefined }"
  >
    <template #prefix>
      <n-icon :color="statusColor === 'success' ? '#18a058' : '#808080'" size="20">
        <component :is="statusIcon" />
      </n-icon>
    </template>

    <div>
      <div style="font-weight: 500;">{{ league.name }}</div>
      <div style="font-size: 0.875rem; opacity: 0.7; display: flex; align-items: center; gap: 8px; margin-top: 4px;">
        <n-tag :type="statusColor === 'success' ? 'success' : 'default'" size="small">
          {{ statusText }}
        </n-tag>
        {{ t('leagues.createdOn') }} {{ formatDate(league.created_at) }}
      </div>
    </div>

    <template #suffix>
      <n-icon color="#808080" size="20">
        <ChevronForwardIcon />
      </n-icon>
    </template>
  </n-list-item>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { NListItem, NIcon, NTag } from 'naive-ui';
import { Trophy as TrophyIcon, Archive as ArchiveIcon, ChevronForward as ChevronForwardIcon } from '@vicons/ionicons5';
import { useI18n } from 'vue-i18n';
import type { League } from '@/api/LeagueApi';

interface Props {
  league: League;
}

const { t, locale } = useI18n();
const props = defineProps<Props>();

const statusColor = computed(() => {
  return props.league.status === 'active' ? 'success' : 'grey';
});

const statusIcon = computed(() => {
  return props.league.status === 'active' ? TrophyIcon : ArchiveIcon;
});

const statusText = computed(() => {
  return props.league.status === 'active' ? t('leagues.active') : t('leagues.archived');
});

const formatDate = (dateStr: string) => {
  const localeMap: Record<string, string> = { 'uk': 'uk-UA', 'en': 'en-US', 'et': 'et-EE' };
  return new Date(dateStr).toLocaleDateString(localeMap[locale.value] || 'en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
};
</script>
