import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import vuetify from './plugins/vuetify'

// Ensure Vuetify styles are imported
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'

createApp(App)
    .use(vuetify)
    .use(router)
    .mount('#app')