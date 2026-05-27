<?php
namespace Bundles\TgBotContext\Model;

use Bundles\TgBotContext\Model\Traits\SendTrait;
use TgBotApi\BotApiBase\BotApiComplete;
use TgBotApi\BotApiBase\Method\SendLocationMethod;

class LocationMessage extends BasicMessage
{
    use SendTrait;

    public function __construct(private float $lat, private float $lon, ?Keyboard $keyboard)
    {
        parent::__construct('', $keyboard);
    }

    /**
     *
     * @throws \Throwable
     */
    #[\Override]
    public function send(BotApiComplete $bot, ConfigurationInterface $configuration): void
    {
        $params   = [];
        $keyboard = $this->getKeyboard();

        if ($keyboard instanceof Keyboard) {
            $params = array_merge([
                'replyMarkup' => $keyboard->getKeyboard(),
            ], $params);
        }

        $method = SendLocationMethod::create(
            $configuration->getChatId(),
            $this->lat,
            $this->lon,
            $params
        );

        $this->handleSend(function () use ($bot, $method) {
            $bot->sendLocation($method);
        });
    }
}
