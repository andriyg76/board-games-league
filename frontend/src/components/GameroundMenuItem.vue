<template>
  <div class="d-none d-md-flex">
    <v-menu>
      <template v-slot:activator="{ props }">
        <v-btn variant="text" v-bind="props">
          {{ t('gameRounds.menu') }}
          <v-icon end>mdi-chevron-down</v-icon>
        </v-btn>
      </template>
      <v-list>
        <v-list-item to="/ui/game-rounds/new">
          <v-list-item-title>{{ t('gameRounds.start') }}</v-list-item-title>
        </v-list-item>
        <v-list-item to="/ui/wizard/new">
          <template v-slot:prepend>
            <v-icon>mdi-wizard-hat</v-icon>
          </template>
          <v-list-item-title>{{ t('wizard.newGame') }}</v-list-item-title>
        </v-list-item>
        <v-divider class="my-1" />
        <v-list-item to="/ui/game-rounds">
          <v-list-item-title>{{ t('gameRounds.list') }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
</script>