<template>
  <div>
    <google-login v-if="!isLoggedIn"/>
    <template v-else>
      <h3>I am a {{ user.name }}</h3>
      <h4>Email: {{ user.email }}</h4>
      <h4>Alias:
        <input v-model="user.alias" @input="checkAliasUniqueness" placeholder="Enter alias"/>
        <span v-if="isAliasUnique">✔️</span>
        <span v-else>❌</span>
      </h4>
      <button @click="updateAlias" :disabled="!isAliasUnique">Update Alias</button>
      <p>
        <img v-if="user.picture" :src="user.picture" :alt="`${user.name} - ${user.email}`"/>
      </p>
    </template>
  </div>
</template>

<script lang="ts" setup>
import {computed, onMounted, ref} from 'vue';
import GoogleLogin from '@/components/GoogleLogin.vue';
import UserApi, { User } from "@/api/UserApi";

const user = ref({} as User)

const isLoggedIn = computed(() => !!user.value.email);

onMounted(async () => {
  try {
    user.value = (await UserApi.getUser()) || {}
  } catch (e) {
    console.error("Error getting user: ", e)
  }
})

const isAliasUnique = ref(false)

async function checkAliasUniqueness() {
  try {
    const response = await UserApi.checkAlias(user.value.alias || "");
    isAliasUnique.value = response.isUnique;
  } catch (e) {
    console.error("Error checking alias uniqueness: ", e);
    isAliasUnique.value = false;
  }
}

const isAliasUpdated = ref(false)

async function updateAlias() {
  if (isAliasUnique.value) {
    try {
      await UserApi.updateUser(user.value);
      isAliasUpdated.value = true;
    } catch (e) {
      console.error("Error updating alias: ", e);
      isAliasUpdated.value = false;
    }
  }
}

</script>