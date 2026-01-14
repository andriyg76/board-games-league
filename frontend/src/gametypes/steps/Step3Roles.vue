<template>
  <v-card flat>
    <v-card-title>{{ $t('game.configureRound') }}</v-card-title>
    <v-card-text>
      <v-text-field
        :model-value="roundName"
        @update:model-value="$emit('update:roundName', $event)"
        :label="$t('game.roundName')"
        variant="outlined"
        class="mb-4"
      />

      <template v-if="roles.length > 0">
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
            <tr v-for="(player, index) in players" :key="player.membership_id">
              <td>
                {{ player.alias }}
                <v-chip v-if="player.isCurrentUser" size="x-small" color="primary" class="ml-1">
                  {{ $t('game.you') }}
                </v-chip>
              </td>
              <td>
                <v-select
                  v-if="assignableRoles.length > 0"
                  :model-value="player.role_key"
                  @update:model-value="$emit('updateRole', index, $event)"
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
                  :model-value="player.is_moderator"
                  @update:model-value="$emit('updateModerator', index, $event)"
                  hide-details
                  density="compact"
                />
              </td>
            </tr>
          </tbody>
        </v-table>
      </template>
    </v-card-text>
    <v-card-actions>
      <v-btn variant="text" @click="$emit('back')" :disabled="saving">
        <v-icon start>mdi-arrow-left</v-icon>
        {{ $t('game.back') }}
      </v-btn>
      <v-spacer />
      <v-btn
        variant="outlined"
        color="primary"
        :loading="saving"
        @click="$emit('save')"
        class="mr-2"
      >
        <v-icon start>mdi-content-save</v-icon>
        {{ $t('game.saveRoles') }}
      </v-btn>
      <v-btn
        color="primary"
        :loading="saving"
        @click="$emit('next')"
      >
        {{ $t('game.next') }}
        <v-icon end>mdi-arrow-right</v-icon>
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import type { Role } from '@/api/GameApi';

export interface RolePlayer {
  membership_id: string;
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
