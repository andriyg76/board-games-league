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
                  <router-link to="/" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('nav.home') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/admin/game-types" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('nav.gameTypes') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/leagues" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('nav.leagues') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/game-rounds" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('gameRounds.title') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/game-rounds/new" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('gameRounds.start') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/game-rounds/new?gameType=wizard" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('wizard.newGame') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/game-rounds" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('gameRounds.list') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="loggedIn" to="/ui/user" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('nav.user') }}
                      </a>
                    </n-list-item>
                  </router-link>
                  <router-link v-if="isSuperAdmin" to="/ui/admin/diagnostics" custom v-slot="{ navigate, href, isActive }">
                    <n-list-item class="mobile-drawer-item" :class="{ 'mobile-drawer-item--active': isActive }">
                      <a :href="href" class="mobile-drawer-link" @click="handleDrawerNavigate(navigate)">
                        {{ t('nav.diagnostics') }}
                      </a>
                    </n-list-item>
                  </router-link>
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
const handleDrawerNavigate = (navigate: (event?: MouseEvent) => void) => {
  navigate();
  drawer.value = false;
};
</script>

<style scoped>
.mobile-drawer-content {
  padding: 8px 0;
}

.mobile-drawer-list {
  font-size: 1.125rem;
}

.mobile-drawer-item {
  padding: 0;
}

.mobile-drawer-item--active .mobile-drawer-link {
  background: rgba(32, 128, 240, 0.12);
  color: #2080f0;
  font-weight: 600;
}

.mobile-drawer-link {
  display: flex;
  align-items: center;
  width: 100%;
  color: inherit;
  text-decoration: none;
  padding: 16px 20px;
  border-radius: 10px;
  transition: background 0.15s ease, color 0.15s ease;
}

.mobile-drawer-link:active {
  background: rgba(32, 128, 240, 0.18);
}

.mobile-drawer-link:focus-visible {
  outline: 2px solid rgba(32, 128, 240, 0.4);
  outline-offset: 2px;
}
</style>
