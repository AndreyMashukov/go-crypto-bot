# Dashboard

> [!Note]
> **Ready to trade?** Go to [autotrade.cloud](https://autotrade.cloud)

**The bot's dashboard is organized into several sections:**
* Portfolio and profit status reports
* Trading stack (a queue table for cryptocurrency purchase orders)
* List of open positions
* Chart block
* List of closed (successful) orders

<img src="/en/img/dashboard/1.png" alt="Dashboard" width="100%"/>

### Portfolio and Profit Status Reports
These reports are located above the trading stack and offer three display options: monthly profit, weekly profit, and daily profit, which can be toggled using a dropdown menu.

### Trading Stack (Queue Table for Cryptocurrency Purchase Orders)
This table arranges cryptocurrencies in a sorted order, offering two modes of sorting: `point` and `percentage`. Points are calculated based on the difference between the calculated purchase price and the current price of the cryptocurrency. Percentages represent the daily change in cryptocurrency prices. Sorting prioritizes larger points downwards and smaller percentages upwards.

Orders are executed strictly in the order they appear in the stack, ensuring funds are first allocated to cryptocurrencies that are likely to be purchased next. This sequencing is crucial because funds are locked on the exchange while limit orders are open.

> [!Note]
> The order of placing limit orders is important to ensure that free funds are first allocated to cryptocurrencies that are most likely to be purchased rather than those whose turn may not come soon. When a limit order is placed on the exchange, funds in the account are locked for the duration of the open limit order.

### Enabling/Disabling Trading Symbols and Purchase Conditions

<img src="/en/img/dashboard/5.png" alt="Enabling/Disabling Trading Symbols and Purchase Conditions" width="100%" style="max-width: 500px;margin-top: 10px;"/>

> [!Note]
> The system allows for quick toggling of trade activation for specific cryptocurrencies directly from the dashboard, facilitated by switches on the trading stack. Additionally, traders can set specific conditions (restrictions) for each cryptocurrency, allowing the bot to execute buying, averaging, or selling actions based on these criteria.

### Configuring Purchase Conditions
With the `BUY IF`, `AVG IF`, and `SELL IF` buttons, users can configure specific conditions for buying, averaging, and selling cryptocurrencies. Conditions can be simple or complex and may be combined using logical operators like `AND` and `OR`.

**Available parameters include:**
* `price` - The current price of the cryptocurrency in USDT
* `daily_percent` - The daily percentage change of the cryptocurrency
* `position_time_minutes` - Time elapsed since the position was opened in minutes
* `extra_orders_today` - Number of averaging orders made today
* `has_signal` - Presence of a trading signal ([AI signal](/en/signal.md) for purchase)

**Available conditions:**
* `=` equals
* `!=` not equal
* `>=` greater than or equal to
* `>` greater than
* `<=` less than or equal to
* `<` less than

Values are entered by the user, and strings are automatically converted into numbers. Care should be taken when setting conditions for negative values, remembering to include the minus sign if necessary.

<img src="/en/img/dashboard/2.png" alt="Condition Configurator for Cryptocurrency Purchasing" width="100%" style="max-width: 400px;margin-top: 10px;"/>

> [!Note]
> **The form above contains 2 conditions and reads as follows:** Purchase `NEOUSDT` if the daily price change percentage for `NEOUSDT` is less than or equal to `-11.00%` OR if `NEOUSDT` has a purchase signal and the daily price change percentage is less than or equal to `-6.00%`. This setup allows the bot to purchase based on specified price triggers or the presence of a trading signal.

### Trading Pair (Symbol) and Prices

<img src="/en/img/dashboard/10.png" alt="Trading Pair (Symbol) and Prices" width="100%" style="max-width: 500px;margin-top: 10px;"/>

In the symbol column, icons next to the name of the trading pair indicate whether trading is disabled or if certain trading conditions are not being met, preventing the bot from purchasing that symbol.

> [!Note]
> To understand which specific trading conditions were not met (either for buying or averaging), check the "Type" column. There are two types of buying operations:
> * `Position Open` - Opening a new position
> * `Extra Charge` - Averaging down on an existing position

Columns following `SYMBOL` in the table include:
* `RATING` - The profitability rating of cryptocurrencies within the `Autotrade.cloud` system, based on closed trades by traders over the last 30 days. Hovering over the icon shows average purchase and selling prices.
* `SIGNALS` - Enabling trading based on AI signals and their configurations, discussed further in the [Trading AI Signals](/en/signal.md) section.
* `PRICE` - The current cost of one unit of cryptocurrency in `USDT`, with a color-coded indicator showing whether the price is below or above the average purchase price.
* `PREDICT` - Short-term price forecast, based on linear regression of order book parameters and historical data over the past 2 days.

> **Starting Trading** - read more in the [Starting Trading](/en/begin.md) section<br>
> **Trading AI Signals** - read more in the [Trading AI Signals](/en/signal.md) section

> [!tip]
> #### Having trouble?
> **Email us at [support@example.com](mailto:support@example.com) and we will help you resolve the issue!**<br>
> **Or join our Telegram chat: [@autotrade_cloud](https://t.me/autotrade_cloud)**
