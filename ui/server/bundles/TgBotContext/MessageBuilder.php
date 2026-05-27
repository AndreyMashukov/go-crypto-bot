<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext;

use Bundles\TgBotContext\Model\BasicMessage;
use Bundles\TgBotContext\Model\Keyboard;
use Bundles\TgBotContext\Model\OrderMessage;

class MessageBuilder
{
    /**
     * @throws \Exception
     */
    public function fromOrderMessage(OrderMessage $orderMessage): array
    {
        $cryptoBot = $orderMessage->getBot();

        $keyboard = (new Keyboard())
            ->addInlineAction('TRADE NOW!', 'https://autotrade.cloud')
            ->toOneRow(1)
            ;

        switch ($orderMessage->getOperation()) {
            case 'SELL':
                $text = "<b>Trader: {$orderMessage->getBot()->getUser()->getNickname()} ({$cryptoBot->getProvider()})</b>\n\n"
                    . "<b>{$orderMessage->getSymbol()}</b>\n"
                    . "Operation: SELL\n"
                    . "Quantity: {$orderMessage->getAmount()}\n"
                    . "Price: {$orderMessage->getPrice()}\n"
                    . $orderMessage->getDetails();
                break;
            case 'BUY':
                $text = "<b>Trader: {$cryptoBot->getUser()->getNickname()} ({$cryptoBot->getProvider()})</b>\n\n"
                    . "<b>{$orderMessage->getSymbol()}</b>\n"
                    . "Operation: BUY\n"
                    . "Quantity: {$orderMessage->getAmount()}\n"
                    . "Price: {$orderMessage->getPrice()}\n"
                    . $orderMessage->getDetails();
                break;
            default:
                throw new \BadMethodCallException('Wrong operation given');
        }

        return [
            new BasicMessage($text, $keyboard),
        ];
    }

    /**
     * @param string $alertText
     *
     * @return BasicMessage[]
     */
    public function getAlertMessages(string $alertText): array
    {
        return [
            new BasicMessage($alertText, null),
        ];
    }
}
