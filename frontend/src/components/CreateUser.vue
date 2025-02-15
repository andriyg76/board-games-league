<template>
  <div>
    <h1>Create User</h1>
    <p v-if="message">{{ message }}</p>
    <div v-if="email">
      <p>Email: {{ email }}</p>
      <button @click="createUser">Create</button>
    </div>
    <div v-else>
      <p>Email parameter is missing</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import UserApi from '../api/UserApi';

const email = ref('');
const message = ref('');
const route = useRoute();

const createUser = async () => {
  try {
    await UserApi.adminCreateUser(email.value);
    message.value = 'User created successfully';
  } catch (error) {
    message.value = 'Failed to create user';
  }
};

onMounted(() => {
  email.value = route.query.email as string;
});
</script>