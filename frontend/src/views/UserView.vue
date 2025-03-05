<template>
  <div>
    {{ userStore.user }}
    {{ userStore.loggedIn }}
    <template v-if="userStore.loggedIn">
      <h3>I am a {{ userStore.user.name }} - {{ userStore.user.external_ids }}</h3>
      <h4>Alias:
        <v-text-field v-model="userStore.user.alias" @input="checkAliasUniqueness" placeholder="Enter alias" append-icon="edit"/>
        <span v-if="isAliasUnique">✔️</span>
        <span v-else>❌</span>
      </h4>
      <v-btn color="primary" @click="updateAlias" :disabled="!isAliasUnique">Update Alias</v-btn>
      <p>
        <v-img v-if="userStore.user.avatar" :src="userStore.user.avatar" :alt="`${userStore.user.name} - ${userStore.user.alias}`" height="64" width="64"/>
      </p>
    </template>
  </div>
</template>

<script lang="ts" setup>
import UserApi from "@/api/UserApi";
import { useUserStore } from '@/store/user';
const userStore = useUserStore();
import {ref} from "vue";

const isAliasUnique = ref(false);

async function checkAliasUniqueness() {
  try {
    const response = await UserApi.checkAlias(userStore.user.alias || "");
    isAliasUnique.value = response.isUnique;
  } catch (e) {
    console.error("Error checking alias uniqueness: ", e);
    isAliasUnique.value = false;
  }
}

const isAliasUpdated = ref(false);

async function updateAlias() {
  if (isAliasUnique.value) {
    try {
      await UserApi.updateUser(userStore.user);
      isAliasUpdated.value = true;
    } catch (e) {
      console.error("Error updating alias: ", e);
      isAliasUpdated.value = false;
    }
  }
}
</script>