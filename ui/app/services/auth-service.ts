import {useCookie} from "#app";
import {Subject} from "rxjs";

export class AuthService {
    private user: any
    public openAuthDialog = new Subject<void>();
    public updateUser = new Subject<void>();

    getToken(): string|null {
      const cookie = useCookie('auth-token');
      return cookie.value || null;
    }

    setToken(token: string): void {
        const cookie = useCookie('auth-token', {
            maxAge: 3600 * 24 * 365,
        });
        cookie.value = token;
    }

    clearToken(): void {
        const cookie = useCookie('auth-token');
        cookie.value = null;
    }

    public setUser(user: any) {
        this.user = user;
    }

    public getUser(): any|null {
        return this.user;
    }

    getEmail(): string | null {
        const cookie = useCookie('user_email');
        return cookie.value || null;
    }

    setEmail(email: string): void {
        const cookie = useCookie('user_email', {
            maxAge: 3600 * 24 * 365,
        });
        cookie.value = email;
    }

    clearEmail(): void {
        const cookie = useCookie('user_email');
        cookie.value = null;
    }

    getPromoCode(): string {
        const cookie = useCookie('promo_code');
        return cookie.value || '';
    }

    setPromoCode(promocode: string): void {
        const cookie = useCookie('promo_code', {
            maxAge: 3600 * 24 * 365,
        });
        cookie.value = promocode;
    }
}
