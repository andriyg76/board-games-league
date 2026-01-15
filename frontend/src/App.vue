<template>
  <n-config-provider>
    <n-message-provider>
      <n-notification-provider>
        <n-dialog-provider>
          <n-layout>
            <n-layout-header bordered style="padding: 12px 16px; display: flex; align-items: center; gap: 8px;">
              <!-- Mobile menu icon -->
              <n-button
                  v-if="isMobile"
                  quaternary
                  @click="drawer = !drawer"
              >
                <template #icon>
                  <n-icon><MenuIcon /></n-icon>
                </template>
              </n-button>

              <!-- Desktop menu -->
              <n-space v-if="!isMobile" :size="8">
                <router-link to="/" style="text-decoration: none;">
                  <n-button quaternary>{{ t('nav.home') }}</n-button>
                </router-link>
                <router-link v-if="loggedIn" to="/ui/admin/game-types" style="text-decoration: none;">
                  <n-button quaternary>{{ t('nav.gameTypes') }}</n-button>
                </router-link>
                <league-menu-item v-if="loggedIn" />
                <gameround-menu-item v-if="loggedIn" />
                <router-link v-if="loggedIn" to="/ui/user" style="text-decoration: none;">
                  <n-button quaternary>{{ t('nav.user') }}</n-button>
                </router-link>
                <router-link v-if="isSuperAdmin" to="/ui/admin/diagnostics" style="text-decoration: none;">
                  <n-button quaternary>{{ t('nav.diagnostics') }}</n-button>
                </router-link>
              </n-space>

              <div style="flex: 1;"></div>
              <n-divider vertical style="margin: 0 8px;" />
              <language-switcher />
              <logout-button/>
            </n-layout-header>

            <!-- Mobile side menu -->
            <n-drawer
                v-model:show="drawer"
                :width="drawerWidth"
                placement="left"
            >
              <n-drawer-content class="mobile-drawer-content" closable>
                <n-list class="mobile-drawer-list">
                  <n-list-item class="mobile-drawer-item">
                    <router-link
                      to="/"
                      class="mobile-drawer-link"
                      @click="drawer = false"
                    >
                      {{ t('nav.home') }}
                    </router-link>
                  </n-list-item>
                  <n-list-item v-if="loggedIn" class="mobile-drawer-item">
                    <router-link
                      to="/ui/admin/game-types"
                      class="mobile-drawer-link"
                      @click="drawer = false"
                    >
                      {{ t('nav.gameTypes') }}
                    </router-link>
                  </n-list-item>
                  <n-list-item v-if="loggedIn" class="mobile-drawer-item">
                    <router-link
                      to="/ui/leagues"
                      class="mobile-drawer-link"
                      @click="drawer = false"
                    >
                      {{ t('nav.leagues') }}
                    </router-link>
                  </n-list-item>
                  <n-list-item v-if="loggedIn" class="mobile-drawer-item">
                    <router-link
                      to="/ui/game-rounds"
                      class="mobile-drawer-link"
                      @click="drawer = false"
                    >
                      {{ t('gameRounds.title') }}
                    </router-link>
                  </n-list-item>
                  <n-list-item v-if="loggedIn" class="mobile-drawer-item">
                    <router-link
                      to="/ui/game-rounds/new"
                      class="mobile-drawer-link"
                      @click="drawer = false"
                    >
                      {{ t('gameRounds.start') }}
                    </router-link>
                  </n-list-item>
                  <n-list-item v-if="loggedIn" class="mobile-drawer-item">
                    <router-link
                      to="/ui/game-rounds/new?gameType=wizard"
                      class="mobile-drawer-link"
                      @click="drawer = false"
                    >
                      {{ t('wizard.newGame') }}
                    </router-link>
                  </n-list-item>
                  <n-list-item v-if="loggedIn" class="mobile-drawer-item">
                    <router-link
                      to="/ui/game-rounds"
                      class="mobile-drawer-link"
                      @click="drawer = false"
                    >
                      {{ t('gameRounds.list') }}
                    </router-link>
                  </n-list-item>
                  <n-list-item v-if="loggedIn" class="mobile-drawer-item">
                    <router-link
                      to="/ui/user"
                      class="mobile-drawer-link"
                      @click="drawer = false"
                    >
                      {{ t('nav.user') }}
                    </router-link>
                  </n-list-item>
                  <n-list-item v-if="isSuperAdmin" class="mobile-drawer-item">
                    <router-link
                      to="/ui/admin/diagnostics"
                      class="mobile-drawer-link"
                      @click="drawer = false"
                    >
                      {{ t('nav.diagnostics') }}
                    </router-link>
                  </n-list-item>
                </n-list>
              </n-drawer-content>
            </n-drawer>

            <n-layout-content style="padding: 24px;">
              <router-view/>
            </n-layout-content>
          </n-layout>
        </n-dialog-provider>
      </n-notification-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { 
  NConfigProvider, 
  NMessageProvider, 
  NNotificationProvider, 
  NDialogProvider,
  NLayout, 
  NLayoutHeader, 
  NLayoutContent, 
  NButton, 
  NSpace, 
  NDivider, 
  NDrawer, 
  NList,
  NListItem,
  NDrawerContent,
  NIcon,
} from 'naive-ui';
import { Menu as MenuIcon } from '@vicons/ionicons5';
import { RouterLink } from 'vue-router';
import LogoutButton from "@/components/LogoutButton.vue";
import { useUserStore } from '@/store/user';
import GameroundMenuItem from "@/components/GameroundMenuItem.vue";
import LeagueMenuItem from "@/components/LeagueMenuItem.vue";
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const userStore = useUserStore();
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'

const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('md')

const drawer = ref(false);
const loggedIn = computed(() => userStore.$state.loggedIn);
const isSuperAdmin = computed(() => userStore.isSuperAdmin);
const drawerWidth = computed(() => (isMobile.value ? '100%' : 320));
</script>

<style scoped>
.mobile-drawer-content {
  padding: 8px 0;
}

.mobile-drawer-list {
  font-size: 1.125rem;
}

.mobile-drawer-item {
  padding: 16px 20px;
}

.mobile-drawer-link {
  display: flex;
  align-items: center;
  width: 100%;
  color: inherit;
  text-decoration: none;
}
</style>
