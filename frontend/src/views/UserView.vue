<template>
  <div>
    <google-login v-if="!userStore.state.loggedIn"/>
    <template v-else>
      <h3>I am a {{ userStore.state.user.name }}</h3>
      <h4>Email: {{ userStore.state.user.email }}</h4>
      <h4>Alias:
        <v-text-field v-model="userStore.state.user.alias" @input="checkAliasUniqueness" placeholder="Enter alias" append-icon="edit"/>
        <span v-if="isAliasUnique">✔️</span>
        <span v-else>❌</span>
      </h4>
      <v-btn color="primary" @click="updateAlias" :disabled="!isAliasUnique">Update Alias</v-btn>
      <p>
        <v-img v-if="userStore.state.user.picture" :src="userStore.state.user.picture" :alt="`${userStore.state.user.name} - ${userStore.state.user.email}`" height="64" width="64"/>
      </p>
    </template>
  </div>
</template>

<script lang="ts" setup>
import GoogleLogin from '@/components/GoogleLogin.vue';
import UserApi from "@/api/UserApi";
import userStore from '@/store/user';

const isAliasUnique = ref(false);

async function checkAliasUniqueness() {
  try {
    const response = await UserApi.checkAlias(userStore.state.user.alias || "");
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
      await UserApi.updateUser(userStore.state.user);
      isAliasUpdated.value = true;
    } catch (e) {
      console.error("Error updating alias: ", e);
      isAliasUpdated.value = false;
    }
  }
}
</script>