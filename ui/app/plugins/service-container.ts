import {AuthService} from '~/services/auth-service';
import {PaymentService} from '~/services/payment-service';
import {EventManager} from '~/services/event-manager';
import {RouterService} from '~/services/router-service';
import {HttpClient} from '~/services/http-client';

export default defineNuxtPlugin(nuxtApp => {
    const router = useRouter()
    const config = useRuntimeConfig();
    const translator: any = nuxtApp.$i18n

    const authService = new AuthService();
    const paymentService = new PaymentService();
    const eventManager = new EventManager();
    const routerService = new RouterService(translator, router);
    const httpClient = new HttpClient(config.public.baseUrl, authService, eventManager);

    return {
        provide: {
            services: {
                authService,
                paymentService,
                eventManager,
                routerService,
                httpClient,
            }
        }
    }
});
