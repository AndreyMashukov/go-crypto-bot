<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

use Bundles\TgBotContext\Model\Traits\SendTrait;
use TgBotApi\BotApiBase\BotApiComplete;
use TgBotApi\BotApiBase\Method\SendPhotoMethod;
use TgBotApi\BotApiBase\Type\InputFileType;

class ImageMessage extends BasicMessage
{
    use SendTrait;

    private string $imageUrl;

    public function __construct(string $message, string $imageUrl, ?Keyboard $keyboard)
    {
        parent::__construct($message, $keyboard);

        $this->imageUrl = $imageUrl;
    }

    /**
     * @param BotApiComplete         $bot
     * @param ConfigurationInterface $configuration
     *
     * @throws \Throwable
     */
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

        $image = $this->imageUrl;

        if (false === mb_stripos($image, 'http')) {
            $image = InputFileType::create($image);
        }

        $method = SendPhotoMethod::create(
            $configuration->getChatId(),
            $image,
            $params
        );

        $this->handleSend(function () use ($bot, $method) {
            $bot->sendPhoto($method);
        });

        // remove file
        if ($image instanceof InputFileType && file_exists($this->imageUrl)) {
            unlink($this->imageUrl);
        }
    }
}
