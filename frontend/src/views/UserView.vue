<template>
  <v-container>
    <template v-if="userStore.loggedIn">
      <h2>Your Profile</h2>
      <v-form @submit.prevent="updateUserProfile">
        <v-row>
          <v-col cols="12" md="6">
            <v-card class="mb-4">
              <v-card-title>Profile Information</v-card-title>
              <v-card-text>
                <p>
                  Current Name: **{{ userStore.user.name }}**
                </p>
                <p>
                  Current Alias: **{{ userStore.user.alias }}**
                  <span v-if="isAliasUnique !== null">
                    <span v-if="isAliasUnique">✔️ (Unique)</span>
                    <span v-else>❌ (Not Unique)</span>
                  </span>
                </p>
                <v-img v-if="userStore.user.avatar" :src="userStore.user.avatar" :alt="`${userStore.user.name}'s avatar`" height="64" width="64" class="my-3"/>
              </v-card-text>
            </v-card>
          </v-col>
          <v-col cols="12" md="6">
            <v-card>
              <v-card-title>Edit Profile</v-card-title>
              <v-card-text>
                <v-text-field
                    v-model="userStore.user.alias"
                    label="Your Alias"
                    @input="checkAliasUniqueness"
                    :rules="[rules.required, rules.aliasUnique]"
                    clearable
                ></v-text-field>

                <v-select
                    v-model="userStore.user.name"
                    :items="userStore.user.names"
                    label="Select Name"
                    class="mt-4"
                ></v-select>

                <v-select
                    v-model="userStore.user.avatar"
                    :items="userStore.user.avatars"
                    label="Select Avatar"
                    class="mt-4"
                >
                  <template #item="{ item }">
                    <v-list-item :key="item.raw"
                                 :prepend-avatar="item.raw"
                    >select</v-list-item>
                  </template>
                  <template #selection="{ item }">
                    <v-avatar :image="item.raw" size="24" class="mr-2"></v-avatar>
                  </template>
                </v-select>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>

        <v-btn
            color="primary"
            type="submit"
            :disabled="!isAliasUnique && userStore.user.alias !== initialAlias"
            class="mt-4"
        >
          Save Profile
        </v-btn>
      </v-form>

      <v-row class="mt-4">
        <v-col cols="12">
          <v-card>
            <v-card-title>Active Sessions</v-card-title>
            <v-card-text>
              <v-skeleton-loader v-if="loadingSessions" type="table"></v-skeleton-loader>
              <v-table v-else-if="sessions.length > 0">
                <thead>
                  <tr>
                    <th>Location</th>
                    <th>IP Address</th>
                    <th>User Agent</th>
                    <th>Created</th>
                    <th>Last Activity</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="session in sessions" :key="session.id" :class="{ 'bg-blue-lighten-5': session.is_current }">
                    <td>
                      <span v-if="session.geo_info">
                        {{ session.geo_info.city || '' }}{{ session.geo_info.city && session.geo_info.country ? ', ' : '' }}{{ session.geo_info.country || '' }}
                      </span>
                      <span v-else class="text-grey">Unknown</span>
                    </td>
                    <td>{{ session.ip_address }}</td>
                    <td class="text-truncate" style="max-width: 300px;">{{ session.user_agent }}</td>
                    <td>{{ formatDate(session.created_at) }}</td>
                    <td>{{ formatDate(session.updated_at) }}</td>
                    <td>
                      <v-chip v-if="session.is_current" color="primary" size="small">Current</v-chip>
                      <v-chip v-else color="default" size="small">Active</v-chip>
                    </td>
                  </tr>
                </tbody>
              </v-table>
              <p v-else>No active sessions found.</p>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </template>
    <template v-else>
      <p>Please log in to view and edit your profile.</p>
    </template>
  </v-container>
</template>

<script lang="ts" setup>
import UserApi, { User, SessionInfo } from "@/api/UserApi";
import { useUserStore } from '@/store/user';
import { ref, onMounted, watch } from "vue";

const userStore = useUserStore();
const isAliasUnique = ref<boolean | null>(null);
const initialAlias = ref<string>('');
const initialName = ref<string>('');
const initialAvatar = ref<string>('');
const sessions = ref<SessionInfo[]>([]);
const loadingSessions = ref(false);

const rules = {
  required: (value: string) => !!value || 'Required.',
  aliasUnique: () => isAliasUnique.value || 'Alias is not unique or too short.',
};

async function checkAliasUniqueness() {
  if (userStore.user.alias && userStore.user.alias.length >= 3) {
    try {
      const response = await UserApi.checkAlias(userStore.user.alias);
      isAliasUnique.value = response.isUnique;
    } catch (e) {
      console.error("Error checking alias uniqueness: ", e);
      isAliasUnique.value = false;
    }
  } else {
    isAliasUnique.value = null; // Reset if alias is too short
  }
}

async function updateUserProfile() {
  // Only update if the alias is unique or hasn't changed from the initial value
  if ((isAliasUnique.value || userStore.user.alias === initialAlias.value) && userStore.user.alias) {
    try {
      await UserApi.updateUser(userStore.user as User);
      console.log('Profile updated successfully!');
      initialAlias.value = userStore.user.alias;
      initialAvatar.value = userStore.user.avatar;
      initialName.value = userStore.user.name;
    } catch (e) {
      console.error("Error updating profile: ", e);
    }
  } else {
    console.warn("Cannot save: Alias is not unique or invalid.");
  }
}

// Watch for changes in the user object's alias to trigger uniqueness check
watch(() => userStore.user.alias, (newAlias, oldAlias) => {
  if (newAlias !== oldAlias && newAlias !== initialAlias.value) {
    checkAliasUniqueness();
  } else if (newAlias === initialAlias.value) {
    isAliasUnique.value = true; // If alias reverts to original, it's unique
  }
});


function formatDate(dateString: string): string {
  if (!dateString) return '';
  const date = new Date(dateString);
  return date.toLocaleString();
}

async function loadSessions() {
  loadingSessions.value = true;
  try {
    // Get rotate token from localStorage if available
    const rotateToken = localStorage.getItem('rotateToken') || undefined;
    sessions.value = await UserApi.getUserSessions(rotateToken || undefined);
  } catch (e) {
    console.error("Error loading sessions:", e);
    sessions.value = [];
  } finally {
    loadingSessions.value = false;
  }
}

onMounted(async () => {
  // Fetch the current user data if not already loaded in the store
  if (!userStore.loggedIn) {
    try {
      const user = await UserApi.getUser();
      if (user) {
        userStore.setUser(user);
      }
    } catch (e) {
      console.error("Error fetching current user:", e);
    }
  }
  initialAlias.value = userStore.user.alias;
  initialName.value = userStore.user.name;
  initialAvatar.value = userStore.user.avatar;
  
  // Load sessions
  if (userStore.loggedIn) {
    await loadSessions();
  }
});
</script>
