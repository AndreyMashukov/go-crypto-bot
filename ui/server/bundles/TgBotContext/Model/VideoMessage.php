<?php
namespace Bundles\TgBotContext\Model;

use Bundles\TgBotContext\Model\Traits\SendTrait;
use TgBotApi\BotApiBase\BotApiComplete;
use TgBotApi\BotApiBase\Method\SendVideoMethod;
use TgBotApi\BotApiBase\Type\InputFileType;

class VideoMessage extends BasicMessage
{
    use SendTrait;

    public function __construct(string $message, private string $videoPath, ?Keyboard $keyboard)
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

        $isHttp = false !== mb_strpos($this->videoPath, 'http');
        $method = SendVideoMethod::create(
            $configuration->getChatId(),
            $isHttp ? $this->videoPath : InputFileType::create($this->videoPath),
            $params
        );

        $this->handleSend(function () use ($bot, $method) {
            $bot->sendVideo($method);
        });
    }
}
