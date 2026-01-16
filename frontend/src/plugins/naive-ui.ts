import { create, NMessageProvider, NNotificationProvider, NDialogProvider, NConfigProvider } from 'naive-ui'

const naive = create({
  components: [
    NMessageProvider,
    NNotificationProvider,
    NDialogProvider,
    NConfigProvider,
  ],
})

export default naive


