import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import vuetify from './plugins/vuetify'
import { i18n } from './i18n'

// Ensure Vuetify styles are imported
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'

createApp(App)
    .use(vuetify)
    .use(router)
    .use(i18n)
    .mount('#app')