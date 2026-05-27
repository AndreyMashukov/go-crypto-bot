<?php
namespace Bundles\TgBotContext\Model;

use Bundles\TgBotContext\Model\Traits\SendTrait;
use TgBotApi\BotApiBase\BotApiComplete;
use TgBotApi\BotApiBase\Method\SendDocumentMethod;

class FileMessage extends BasicMessage
{
    use SendTrait;

    public function __construct(string $message, private string $fileUrl, ?Keyboard $keyboard)
    {
        parent::__construct($message, $keyboard);
    }

    /**
     *
     * @throws \Throwable
     */
    #[\Override]
    public function send(BotApiComplete $bot, ConfigurationInterface $configuration): void
    {
        $params = [
            'parseMode' => 'html',
            'caption'   => $this->getMessage(),
        ];

        $keyboard = $this->getKeyboard();

        if ($keyboard instanceof Keyboard) {
            $params = array_merge([
                'replyMarkup' => $keyboard->getKeyboard(),
            ], $params);
        }

        $method = SendDocumentMethod::create(
            $configuration->getChatId(),
            $this->fileUrl,
            $params
        );

        $this->handleSend(function () use ($bot, $method) {
            $bot->sendDocument($method);
        });
    }
}
