import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import colors from 'vuetify/lib/util/colors.mjs'

export default defineNuxtPlugin((app) => {
    const vuetify = createVuetify({
        theme: {
            themes: {
                light: {
                    dark: false,
                    colors: {
                        primary: colors.teal.darken4,
                        secondary: colors.teal.accent3,
                        success: colors.teal.accent3,
                        warning: colors.lime.lighten3,
                    },
                },
            },
        },
    })
    app.vueApp.use(vuetify)
})
