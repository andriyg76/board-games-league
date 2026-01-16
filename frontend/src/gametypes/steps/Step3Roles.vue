<template>
  <n-card>
    <template #header>
      {{ $t('game.configureRound') }}
    </template>
    <n-input
      :value="roundName"
      @update:value="$emit('update:roundName', $event)"
      :placeholder="$t('game.roundName')"
      style="margin-bottom: 16px;"
      data-testid="round-name-input"
    />

    <template v-if="roles.length > 0">
      <h4 style="margin-bottom: 8px; font-size: 1rem; font-weight: 500;">{{ $t('game.assignRoles') }}</h4>
      <n-table>
        <thead>
          <tr>
            <th>{{ $t('leagues.player') }}</th>
            <th>{{ $t('gameTypes.roles') }}</th>
            <th v-if="hasModerator">{{ $t('roleTypes.moderator') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(player, index) in players" :key="player.membership_code || player.membership_id">
            <td>
              <div style="display: flex; align-items: center; gap: 8px;">
                <span>{{ player.alias }}</span>
                <n-tag v-if="player.isCurrentUser" size="small" type="primary">
                  {{ $t('game.you') }}
                </n-tag>
              </div>
            </td>
            <td>
              <n-select
                v-if="assignableRoles.length > 0"
                :value="player.role_key"
                @update:value="$emit('updateRole', index, $event)"
                :options="assignableRoles.map(r => ({ label: r.name, value: r.key }))"
                placeholder="-"
                size="small"
                style="min-width: 150px;"
              />
              <span v-else style="color: #999;">-</span>
            </td>
            <td v-if="hasModerator">
              <n-checkbox
                :checked="player.is_moderator"
                @update:checked="$emit('updateModerator', index, !!$event)"
              />
            </td>
          </tr>
        </tbody>
      </n-table>
    </template>
    <template #action>
      <div style="display: flex; justify-content: space-between; align-items: center; padding: 16px;">
        <n-button quaternary @click="$emit('back')" :disabled="saving">
          <template #icon>
            <n-icon><ChevronBackIcon /></n-icon>
          </template>
          {{ $t('game.back') }}
        </n-button>
        <div style="display: flex; gap: 8px;">
          <n-button
            secondary
            type="primary"
            :loading="saving"
            @click="$emit('save')"
          >
            <template #icon>
              <n-icon><SaveIcon /></n-icon>
            </template>
            {{ $t('game.saveRoles') }}
          </n-button>
          <n-button
            type="primary"
            :loading="saving"
            data-testid="configure-round-next-button"
            @click="$emit('next')"
          >
            {{ $t('game.next') }}
            <template #icon>
              <n-icon><ChevronForwardIcon /></n-icon>
            </template>
          </n-button>
        </div>
      </div>
    </template>
  </n-card>
</template>

<script lang="ts" setup>
import { NCard, NInput, NTable, NTag, NSelect, NCheckbox, NButton, NIcon } from 'naive-ui';
import { ChevronBack as ChevronBackIcon, ChevronForward as ChevronForwardIcon, Save as SaveIcon } from '@vicons/ionicons5';
import type { Role } from '@/api/GameApi';

export interface RolePlayer {
  membership_code: string;
  membership_id?: string; // Legacy support
  alias: string;
  is_moderator: boolean;
  role_key?: string;
  isCurrentUser?: boolean;
}

export interface AssignableRole {
  key: string;
  name: string;
}

defineProps<{
  roundName: string;
  players: RolePlayer[];
  roles: Role[];
  assignableRoles: AssignableRole[];
  hasModerator: boolean;
  saving: boolean;
}>();

defineEmits<{
  'update:roundName': [name: string];
  updateRole: [index: number, roleKey: string];
  updateModerator: [index: number, isModerator: boolean];
  back: [];
  save: [];
  next: [];
}>();
</script>
