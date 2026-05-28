import type {VueI18n} from "vue-i18n";
import type {Router} from '#vue-router';

export class RouterService {
  constructor(private i18n: VueI18n, private router: Router) {}

  public navigate(path: string): Promise<any> {
    let prefix = '';

    if (this.i18n.locale._value !== 'en') {
      prefix = `/${this.i18n.locale._value}`
    }

    return new Promise((resolve) => {
        setTimeout(() => {
            window.location.href = `${prefix}${path}`;
        }, 150)
        resolve('done')
    })
  }
}
