<template>
  <div>
    <n-alert v-if="standings.length === 0" type="info" style="margin-bottom: 16px;">
      {{ t('leagues.noLeagueData') }}
    </n-alert>

    <n-data-table
      v-else
      :columns="columns"
      :data="standings"
      :pagination="{ pageSize: 10 }"
    />

    <!-- Details Dialog -->
    <n-modal v-model:show="detailsDialog" preset="card" :title="selectedPlayer?.user_name" style="max-width: 600px;">
      <template v-if="selectedPlayer">
        <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 16px;">
          <n-avatar :src="selectedPlayer.user_avatar" :size="48" round />
          <div>
            <div style="font-weight: 500;">{{ selectedPlayer.user_name }}</div>
            <div style="font-size: 0.875rem; opacity: 0.7;">{{ t('leagues.detailedStats') }}</div>
          </div>
        </div>

        <n-divider />

        <n-grid :cols="24" :x-gap="16" style="margin-bottom: 16px;">
          <n-gi :span="12">
            <n-card style="text-align: center; background: rgba(32, 128, 240, 0.1);">
              <div style="font-size: 2rem; font-weight: 500;">{{ selectedPlayer.total_points }}</div>
              <div style="font-size: 0.875rem; opacity: 0.7;">{{ t('leagues.totalPoints') }}</div>
            </n-card>
          </n-gi>
          <n-gi :span="12">
            <n-card style="text-align: center; background: rgba(24, 160, 88, 0.1);">
              <div style="font-size: 2rem; font-weight: 500;">{{ selectedPlayer.games_played }}</div>
              <div style="font-size: 0.875rem; opacity: 0.7;">{{ t('leagues.gamesPlayed') }}</div>
            </n-card>
          </n-gi>
        </n-grid>

        <n-divider style="margin: 16px 0;" />

        <div style="font-size: 0.875rem; font-weight: 500; margin-bottom: 8px;">{{ t('leagues.pointsBreakdown') }}</div>
        <n-list>
          <n-list-item>
            <template #prefix>
              <n-icon color="#18a058"><PersonCheckIcon /></n-icon>
            </template>
            <div>{{ t('leagues.participationPoints') }}</div>
            <template #suffix>
              <n-tag size="small">{{ selectedPlayer.participation_points }}</n-tag>
            </template>
          </n-list-item>
          <n-list-item>
            <template #prefix>
              <n-icon color="#f0a020"><TrophyIcon /></n-icon>
            </template>
            <div>{{ t('leagues.positionPoints') }}</div>
            <template #suffix>
              <n-tag size="small">{{ selectedPlayer.position_points }}</n-tag>
            </template>
          </n-list-item>
          <n-list-item>
            <template #prefix>
              <n-icon color="#2080f0"><GavelIcon /></n-icon>
            </template>
            <div>{{ t('leagues.moderationPoints') }}</div>
            <template #suffix>
              <n-tag size="small">{{ selectedPlayer.moderation_points }}</n-tag>
            </template>
          </n-list-item>
        </n-list>

        <n-divider style="margin: 16px 0;" />

        <div style="font-size: 0.875rem; font-weight: 500; margin-bottom: 8px;">{{ t('leagues.podiums') }}:</div>
        <n-grid :cols="24" :x-gap="16">
          <n-gi :span="8">
            <n-card style="text-align: center; background: rgba(255, 215, 0, 0.1);">
              <n-icon size="32" color="#ffd700" style="margin-bottom: 8px;">
                <TrophyIcon />
              </n-icon>
              <div style="font-size: 1.25rem; font-weight: 500;">{{ selectedPlayer.first_place_count }}</div>
              <div style="font-size: 0.875rem; opacity: 0.7;">{{ t('leagues.firstPlace') }}</div>
            </n-card>
          </n-gi>
          <n-gi :span="8">
            <n-card style="text-align: center; background: rgba(192, 192, 192, 0.1);">
              <n-icon size="32" color="#c0c0c0" style="margin-bottom: 8px;">
                <TrophyIcon />
              </n-icon>
              <div style="font-size: 1.25rem; font-weight: 500;">{{ selectedPlayer.second_place_count }}</div>
              <div style="font-size: 0.875rem; opacity: 0.7;">{{ t('leagues.secondPlace') }}</div>
            </n-card>
          </n-gi>
          <n-gi :span="8">
            <n-card style="text-align: center; background: rgba(205, 127, 50, 0.1);">
              <n-icon size="32" color="#cd7f32" style="margin-bottom: 8px;">
                <TrophyIcon />
              </n-icon>
              <div style="font-size: 1.25rem; font-weight: 500;">{{ selectedPlayer.third_place_count }}</div>
              <div style="font-size: 0.875rem; opacity: 0.7;">{{ t('leagues.thirdPlace') }}</div>
            </n-card>
          </n-gi>
        </n-grid>

        <n-divider style="margin: 16px 0;" />

        <n-list>
          <n-list-item>
            <div>{{ t('leagues.gamesAsModerator') }}</div>
            <template #suffix>
              <n-tag size="small">{{ selectedPlayer.games_moderated }}</n-tag>
            </template>
          </n-list-item>
        </n-list>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, h } from 'vue';
import { NAlert, NDataTable, NModal, NAvatar, NDivider, NGrid, NGi, NCard, NIcon, NList, NListItem, NTag, NButton, DataTableColumns } from 'naive-ui';
import { Medal as MedalIcon, Trophy as TrophyIcon, Information as InformationIcon, CheckmarkCircle as PersonCheckIcon, Hammer as GavelIcon } from '@vicons/ionicons5';
import { useI18n } from 'vue-i18n';
import type { LeagueStanding } from '@/api/LeagueApi';

interface Props {
  standings: LeagueStanding[];
}

const { t } = useI18n();
const props = defineProps<Props>();

const columns = computed<DataTableColumns<LeagueStanding>>(() => [
  { 
    title: '#', 
    key: 'position', 
    width: 80,
    render: (_row: LeagueStanding, index: number) => {
      const position = index + 1;
      let icon = null;
      let color = '';
      
      if (index === 0) {
        icon = h(MedalIcon);
        color = '#ffd700';
      } else if (index === 1) {
        icon = h(MedalIcon);
        color = '#c0c0c0';
      } else if (index === 2) {
        icon = h(MedalIcon);
        color = '#cd7f32';
      }
      
      return h('div', { style: 'display: flex; align-items: center; gap: 8px;' }, [
        icon ? h(NIcon, { color, size: 16 }, { default: () => icon }) : null,
        h('span', position.toString())
      ]);
    }
  },
  { 
    title: t('leagues.player'), 
    key: 'user',
    render: (row: LeagueStanding) => {
      return h('div', { style: 'display: flex; align-items: center; gap: 12px; padding: 8px 0;' }, [
        h(NAvatar, { src: row.user_avatar, size: 32, round: true }),
        h('span', row.user_name)
      ]);
    }
  },
  { 
    title: t('leagues.points'), 
    key: 'total_points',
    align: 'center',
    render: (row: LeagueStanding) => {
      return h(NTag, { type: 'primary' }, { default: () => row.total_points.toString() });
    }
  },
  { 
    title: t('leagues.games'), 
    key: 'games_played',
    align: 'center'
  },
  { 
    title: t('leagues.podiums'), 
    key: 'podiums',
    align: 'center',
    render: (row: LeagueStanding) => {
      return h('div', { style: 'display: flex; gap: 4px; justify-content: center;' }, [
        row.first_place_count > 0 ? h(NTag, { style: 'background: #ffd700; color: #000;' }, { 
          default: () => h('div', { style: 'display: flex; align-items: center; gap: 4px;' }, [
            h(NIcon, { size: 14 }, { default: () => h(TrophyIcon) }),
            h('span', row.first_place_count.toString())
          ])
        }) : null,
        row.second_place_count > 0 ? h(NTag, { style: 'background: #c0c0c0; color: #000;' }, { 
          default: () => h('div', { style: 'display: flex; align-items: center; gap: 4px;' }, [
            h(NIcon, { size: 14 }, { default: () => h(TrophyIcon) }),
            h('span', row.second_place_count.toString())
          ])
        }) : null,
        row.third_place_count > 0 ? h(NTag, { style: 'background: #cd7f32; color: #fff;' }, { 
          default: () => h('div', { style: 'display: flex; align-items: center; gap: 4px;' }, [
            h(NIcon, { size: 14 }, { default: () => h(TrophyIcon) }),
            h('span', row.third_place_count.toString())
          ])
        }) : null,
      ]);
    }
  },
  { 
    title: '', 
    key: 'details',
    width: 60,
    render: (row: LeagueStanding) => {
      return h(NButton, { 
        quaternary: true,
        circle: true,
        size: 'small',
        onClick: () => showDetails(row)
      }, {
        icon: () => h(NIcon, { size: 16 }, { default: () => h(InformationIcon) })
      });
    }
  },
]);

const detailsDialog = ref(false);
const selectedPlayer = ref<LeagueStanding | null>(null);

const showDetails = (player: LeagueStanding) => {
  selectedPlayer.value = player;
  detailsDialog.value = true;
};
</script>
