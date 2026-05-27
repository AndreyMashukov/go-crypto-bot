export class Confirmation {
    title: string;
    text: string;
    closeBtnText: string;
    okBtnText: string;
    closeCallback: () => void;
    okCallback: () => void;

    constructor(
        title: string,
        text: string,
        closeBtnText: string,
        okBtnText: string,
        closeCallback: () => void,
        okCallback: () => void
    ) {
        this.title = title;
        this.text = text;
        this.closeBtnText = closeBtnText;
        this.okBtnText = okBtnText;
        this.closeCallback = closeCallback;
        this.okCallback = okCallback;
    }
}