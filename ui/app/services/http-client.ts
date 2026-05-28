import type {AuthService} from '~/services/auth-service';
import {Alert} from '~/model/alert';
import {AlertEvent} from '~/model/alert-event';
import type {EventManager} from '~/services/event-manager';
import type {HTTPMethod} from 'h3';
import {rejects} from 'node:assert';

export class HttpClient {
    constructor(private baseUrl: string, private authService: AuthService, private eventManager: EventManager) {
    }

    public secureGetServer(url: string): Promise<any|Error> {
        return this.request('GET', url, true, null, true);
    }

    public publicGetClient(url: string): Promise<any|Error> {
        return this.request('GET', url, false, null, false);
    }

    public secureGetClient(url: string): Promise<any|Error> {
        return this.request('GET', url, false, null, true);
    }

    public securePutClient(url: string, body: any = null): Promise<any|Error> {
        return this.request('PUT', url, false, body, true);
    }

    public securePatchClient(url: string, body: any = null): Promise<any|Error> {
        return this.request('PATCH', url, false, body, true);
    }

    public securePostClient(url: string, body: any = null): Promise<any|Error> {
        return this.request('POST', url, false, body, true);
    }

    public secureDeleteClient(url: string): Promise<any|Error> {
        return this.request('DELETE', url, false, null, true);
    }

    public request(method: HTTPMethod, url: string, server: boolean, body: any, secure: boolean): Promise<any|Error> {
        return new Promise((resolve, reject) => {
            let headers = {}

            if (secure) {
                const token = this.authService.getToken()

                if (!token) {
                    reject(new Error('Authorization required'));
                    return;
                }

                headers = {
                    Authorization: `Bearer ${token}`,
                };
            }

            useFetch(url, {
                method: method,
                baseURL: this.baseUrl,
                server,
                body,
                headers,
            }).then(({data: data, error: error}: any) => {
                if (error.value) {
                    let message = error.value.data

                    if (error.value.data && error.value.data.message) {
                        message = error.value.data.message
                    }

                    if (message) {
                        const alertType = Alert.TYPE_ERROR;
                        this.eventManager.alert(
                            new AlertEvent(
                                new Alert(message, alertType),
                                4000
                            )
                        );
                    }
                    if (!message) {
                        message = 'Bad request.';
                    }
                    reject(new Error(message));
                    return;
                }

                resolve(data.value);
            }).catch((err) => {
                console.error(err);
                reject(new Error('Something went wrong'))
            });
        })
    }
}