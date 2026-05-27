import {VueI18n} from "vue-i18n";
import {Router} from '#vue-router';

export class RouterService {
  constructor(private i18n: VueI18n, private router: Router) {}

  public navigate(path: string): Promise<any> {
    let prefix = '';

    if (this.i18n.locale._value !== 'en') {
      prefix = `/${this.i18n.locale._value}`
    }

    // Nuxt router works wrong and we get error during page render, this is temporary solution.
    return new Promise((resolve) => {
        setTimeout(() => {
            window.location.href = `${prefix}${path}`;
        }, 150)
        resolve('done')
    })
  }
}
