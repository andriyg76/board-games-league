<template>
  <v-app>
    <v-app-bar app>
      <!-- Mobile menu icon -->
      <v-app-bar-nav-icon
          @click="drawer = !drawer"
          class="d-flex d-md-none"
      ></v-app-bar-nav-icon>

      <!-- Desktop menu -->
      <div class="d-none d-md-flex">
        <v-btn to="/" variant="text">{{ t('nav.home') }}</v-btn>
        <v-btn
            to="/ui/admin/game-types"
            v-if="loggedIn"
            variant="text"
        >{{ t('nav.gameTypes') }}</v-btn>
        <v-btn
            to="/ui/leagues"
            v-if="loggedIn"
            variant="text"
        >{{ t('nav.leagues') }}</v-btn>
        <gameround-menu-item v-if="loggedIn" />
        <v-btn
            to="/ui/user"
            v-if="loggedIn"
            variant="text"
        >{{ t('nav.user') }}</v-btn>
      </div>

      <v-spacer></v-spacer>
      <v-divider vertical class="mx-2"></v-divider>
      <language-switcher />
      <logout-button/>
    </v-app-bar>

    <!-- Mobile side menu -->
    <v-navigation-drawer
        v-model="drawer"
        temporary
        class="d-md-none"
    >
      <v-list>
        <v-list-item to="/" :title="t('nav.home')" />
        <v-list-item
            v-if="loggedIn"
            to="/ui/admin/game-types"
            :title="t('nav.gameTypes')"
        />
        <v-list-item
            v-if="loggedIn"
            to="/ui/leagues"
            :title="t('nav.leagues')"
        />
        <v-list-item
            v-if="loggedIn"
            to="/ui/game-rounds"
            :title="t('gameRounds.title')"
        />
        <v-list-group v-if="loggedIn" value="Game Rounds">
          <v-list-item to="/ui/game-rounds/new" :title="t('gameRounds.start')" />
          <v-list-item to="/ui/game-rounds" :title="t('gameRounds.list')" />
        </v-list-group>
        <v-list-item
            v-if="loggedIn"
            to="/ui/user"
            :title="t('nav.user')"
        />
      </v-list>
    </v-navigation-drawer>

    <v-main>
      <v-container>
        <router-view/>
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import LogoutButton from "@/components/LogoutButton.vue";
import {defineComponent, computed, ref} from "vue";
import { useUserStore } from '@/store/user';
const userStore = useUserStore();
import GameroundMenuItem from "@/components/GameroundMenuItem.vue";
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import {useI18n} from "vue-i18n";

const { t } = useI18n();

const drawer = ref(false);
const loggedIn = computed(() => userStore.$state.loggedIn);

defineComponent({
  components: {
    LogoutButton,
    LanguageSwitcher,
    GameroundMenuItem,
  },
});
</script>
