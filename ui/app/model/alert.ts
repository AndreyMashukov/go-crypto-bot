export class Alert {
  public static readonly TYPE_ERROR   = 'error';
  public static readonly TYPE_SUCCESS = 'success'

  constructor(public text: string, public type: string) {
  }

  public getColor(): string {
    switch (this.type) {
      case Alert.TYPE_SUCCESS:
        return 'success';
      case Alert.TYPE_ERROR:
        return 'red accent-2';
      default:
        return 'default';
    }
  }
}
