<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

use Bundles\TgBotContext\Model\Traits\SendTrait;
use TgBotApi\BotApiBase\BotApiComplete;
use TgBotApi\BotApiBase\Method\SendLocationMethod;

class LocationMessage extends BasicMessage
{
    use SendTrait;

    private float $lat;

    private float $lon;

    public function __construct(float $lat, float $lon, ?Keyboard $keyboard)
    {
        parent::__construct('', $keyboard);

        $this->lat = $lat;
        $this->lon = $lon;
    }

    /**
     * @param BotApiComplete         $bot
     * @param ConfigurationInterface $configuration
     *
     * @throws \Throwable
     */
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
