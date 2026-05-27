# Starting Trading

> [!Note]
> **Ready to trade?** Go to [autotrade.cloud](https://autotrade.cloud)

<img src="/en/img/begin/1.png" alt="Trade Pair Settings" width="100%"/>

On the bot setup page, in addition to API key settings, below you will find the configuration for trading pairs. You can set up the cryptocurrencies you plan to trade here, although they can also be adjusted on the [Dashboard](dashboard.md). However, the dashboard does not allow you to set target profits and tiered averaging, which is the main difference.

> [!Tip]
> We recommend creating trading pairs in "disabled" state and enabling them later on the [Dashboard](/en/dashboard.md) after setting trading conditions and finalizing your trading strategy.

### Let's go through the elements on this page:
* **Enabled (Buy)** - Toggle for the trading pair's state, we recommend keeping it off and turning it on later on the [Dashboard](/en/dashboard.md)
* **Symbol** - Trading pair, such as `BTCUSDT` or `ETHUSDT` (the cryptocurrency the bot will trade)
* **Budget** - The amount in USDT for opening a position when placing a limit order to buy cryptocurrency. This amount will be used by the bot to buy the cryptocurrency at the position opening, averaging is set separately. (do not specify the entire deposit here, we recommend using only part of the deposit for the initial market entry, leave a reserve for averaging)
* **Buy/Sell** - Button that opens a dialog box for configuring profit and averaging

### Profit Configuration
<img src="/en/img/begin/2.png" alt="Profit Configuration in the Trading Bot" width="100%" style="max-width: 700px;"/>

The tiered profit configuration allows setting a dynamic profit percentage depending on the time elapsed since the position opening (the purchase of cryptocurrency by the bot on the exchange)

**For example, as shown in the screenshot above:**
* 3 minutes after the order opens, the profit percentage is +15%
* 8 minutes after the order opens, the profit percentage is +10%
* 15 minutes after the order opens, the profit percentage is +5%

> [!Note]
> Options do not consider the previous step, i.e., the time does not accumulate but is counted from the moment of position opening. The first option is active for the first 3 minutes, then the second is activated and active for another 5 minutes (8 minutes after opening), then the third is activated which remains active until the position closes (because there are no options after it, it is the last)

### Averaging Configuration
<img src="/en/img/begin/3.png" alt="Averaging Configuration in the Trading Bot" width="100%" style="max-width: 600px;"/>

A similar tiered configuration for averaging positions allows buying more cryptocurrency as the price drops, aiming to reduce the "average" price and exit the deal earlier. This practice might seem risky, but in reality, it allows for an earlier exit from the deal than waiting for a rise without averaging poorly purchased positions.

**For example, as shown in the screenshot above:**
* If the profit percentage drops to -10%, the bot will buy cryptocurrency worth 80 USDT
* If the profit percentage drops to -20%, the bot will buy cryptocurrency worth 100 USDT
* If the profit percentage drops to -30%, the bot will buy cryptocurrency worth 200 USDT

> [!Note]
> **How does averaging work?** Imagine you bought a watermelon for 100 rubles and want to sell it for more, say for 110 rubles (+10%), thus making a profit of 10 rubles. But if the market price for watermelon falls to 80 rubles (-20%), and you buy a second watermelon for 80 rubles, then the average price of your watermelons becomes 90 rubles. When the price returns to 100 rubles, you can sell both watermelons and earn 20 rubles profit (about +10%). This is how averaging works in trading.

[//]: # ( todo: about restarting the bot after changing configurations )

> **Binance Bot Setup** - read more in the section [Binance Bot Setup](/en/binance.md)<br>
> **ByBit Bot Setup** - read more in the section [ByBit Bot Setup](/en/bybit.md)<br>
> **Dashboard** - read more in the section [Dashboard](/en/dashboard.md)

> [!tip]
> #### Having trouble?
> **Email us at [support@example.com](mailto:support@example.com) and we will help you resolve the issue!**<br>
> **Or join our Telegram chat: [@autotrade_cloud](https://t.me/autotrade_cloud)**
