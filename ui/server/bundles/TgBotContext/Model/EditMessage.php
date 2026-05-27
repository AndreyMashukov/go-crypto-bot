<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

use Bundles\TgBotContext\Model\Traits\SendTrait;
use TgBotApi\BotApiBase\BotApiComplete;
use TgBotApi\BotApiBase\Exception\ResponseException;
use TgBotApi\BotApiBase\Method\EditMessageTextMethod;

class EditMessage extends BasicMessage
{
    use SendTrait;

    private int $messageId;

    public function __construct(string $messageText, int $messageId, ?Keyboard $keyboard)
    {
        parent::__construct($messageText, $keyboard);

        $this->messageId = $messageId;
    }

    public function send(BotApiComplete $bot, ConfigurationInterface $configuration): void
    {
        $params   = ['parseMode' => 'html'];
        $keyboard = $this->getKeyboard();

        if ($keyboard instanceof Keyboard) {
            $params = array_merge([
                'replyMarkup' => $keyboard->getKeyboard(),
            ], $params);
        }

        $message = EditMessageTextMethod::create(
            $configuration->getChatId(),
            $this->messageId,
            $this->getMessage(),
            $params
        );

        try {
            $this->handleSend(function () use ($bot, $message) {
                $bot->editMessageText($message);
            });
        } catch (ResponseException $exception) {
            if (false === mb_strpos($exception->getMessage(), 'message is not modified')) {
                throw $exception;
            }
        }
    }
}
