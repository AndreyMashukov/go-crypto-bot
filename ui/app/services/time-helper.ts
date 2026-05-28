export function stringToDate(value: string): Date {
    return new Date(value.replace(/-/g, '/').replace('T', ' ').replace(/(\..*|\+.*)/, ""));
}

export function fix(value: number): string {
    if (value < 10) {
        return `0${value}`;
    }

    return value.toString();
}

export function getFormatted(value: string): string {
    const date  = stringToDate(value);
    const day   = fix(date.getDate());
    const month = fix(date.getMonth() + 1);

    const hours   = fix(date.getHours());
    const minutes = fix(date.getMinutes());
    const seconds = fix(date.getSeconds());

    return `${day}.${month}.${date.getFullYear()} ${hours}:${minutes}:${seconds}`;
}

export function getShortFormatted(value: string): string {
    const date  = stringToDate(value);
    const month = fix(date.getMonth() + 1);

    return `${month}/${date.getFullYear().toString().slice(-2)}`;
}
