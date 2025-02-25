<template>
  <v-app>
    <v-app-bar app>
      <v-btn to="/">Home</v-btn>
      <v-btn to="/ui/admin/game-types" v-if="loggedIn">Game Types</v-btn>
      <v-btn to="/ui/user" v-if="loggedIn">User</v-btn>
      <v-spacer></v-spacer>
      <language-switcher />
      <logout-button/>
    </v-app-bar>
    <v-main>
      <v-container>

        <router-view/>
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import LogoutButton from "@/components/LogoutButton.vue";
import {defineComponent, ref} from "vue";
import userStore from "@/store/user"
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";

const loggedIn = ref(userStore.state)

defineComponent({
  components: {
    LogoutButton,
    LanguageSwitcher,
  },
})

</script>

<style>
</style>
