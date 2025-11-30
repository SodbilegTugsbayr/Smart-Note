import '@mdi/font/css/materialdesignicons.css'
import "vuetify/styles";
import { createVuetify, type ThemeDefinition } from "vuetify";

export default defineNuxtPlugin((app) => {
  const customTheme: ThemeDefinition = {
    dark: false,
    colors: {
      primary: "#1867c0",
    }
  };
  const vuetify = createVuetify({
    defaults: {
      VBtn: {
        // style: "text-transform: none;" // Modify if needed
      }
    },
    theme: {
      defaultTheme: "customTheme",
      themes: {
        customTheme
      }
    },
  });
  app.vueApp.use(vuetify);
});
