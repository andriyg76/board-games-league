<template>
  <v-app>
    <v-app-bar app>
      <v-toolbar-title>My App</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn to="/">Home</v-btn>
      <v-btn to="/ui/admin/game-types">Game Types</v-btn>
      <v-btn to="/ui/user">User</v-btn>
      <v-spacer></v-spacer>
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
import {defineComponent} from "vue";

defineComponent({
  components: {
    LogoutButton
  },
})

</script>

<style>
</style>
