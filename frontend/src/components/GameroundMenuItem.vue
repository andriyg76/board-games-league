<template>
  <n-dropdown :options="menuOptions" trigger="hover" @select="handleSelect">
    <n-button quaternary>
      {{ t('gameRounds.menu') }}
      <template #icon>
        <n-icon><ChevronDownIcon /></n-icon>
      </template>
    </n-button>
  </n-dropdown>
</template>

<script lang="ts" setup>
import { h, computed } from 'vue';
import { NButton, NDropdown, NIcon, useMessage } from 'naive-ui';
import { ChevronDown as ChevronDownIcon } from '@vicons/ionicons5';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

const { t } = useI18n();
const router = useRouter();
const message = useMessage();

const menuOptions = computed(() => [
  {
    label: t('gameRounds.start'),
    key: 'new',
  },
  {
    label: t('wizard.newGame'),
    key: 'wizard',
  },
  {
    type: 'divider',
    key: 'divider',
  },
  {
    label: t('gameRounds.list'),
    key: 'list',
  },
]);

const handleSelect = (key: string) => {
  if (key === 'new') {
    router.push('/ui/game-rounds/new');
  } else if (key === 'wizard') {
    router.push('/ui/game-rounds/new?gameType=wizard');
  } else if (key === 'list') {
    router.push('/ui/game-rounds');
  }
};
</script>