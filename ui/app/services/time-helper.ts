export class TimeHelper {
    static getUtcTime(): number {
        let date    = new Date();
        let utcTime = Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate(), date.getUTCHours(), date.getUTCMinutes(), date.getUTCSeconds(), date.getUTCMilliseconds());

        return Math.ceil(utcTime / 1000);
    }

    static stringToDate(string: string): Date {
        return new Date(string.replace(/-/g,'/').replace('T',' ').replace(/(\..*|\+.*)/,""));
    }

    static getFormatted(string: string): string {
        const date  = TimeHelper.stringToDate(string);
        const day   = TimeHelper.fix(date.getDate());
        const month = TimeHelper.fix(date.getMonth() + 1);

        const hours   = TimeHelper.fix(date.getHours());
        const minutes = TimeHelper.fix(date.getMinutes());
        const seconds = TimeHelper.fix(date.getSeconds());

        return `${day}.${month}.${date.getFullYear()} ${hours}:${minutes}:${seconds}`
    }

    static getShortFormatted(string: string): string {
        const date  = TimeHelper.stringToDate(string);
        const month = TimeHelper.fix(date.getMonth() + 1);

        return `${month}/${date.getFullYear().toString().substr(-2)}`
    }

    static fix(value) {
        if (value < 10) {
            return `0${value}`;
        }

        return value;
    }
}
