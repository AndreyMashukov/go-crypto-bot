<?php
namespace Bundles\TgBotContext;

use Bundles\TgBotContext\Exception\SkipStrategyException;
use Bundles\TgBotContext\Model\BasicMessage;
use Bundles\TgBotContext\Model\ConfigurationInterface;
use Bundles\TgBotContext\Model\Input;
use Bundles\TgBotContext\Strategy\StrategyInterface;
use Psr\Log\LoggerInterface;
use TgBotApi\BotApiBase\BotApiComplete;
use TgBotApi\BotApiBase\Exception as TGException;
use TgBotApi\BotApiBase\Method\SendMessageMethod;

class Context
{
    private array $strategies;

    private readonly BotApiComplete $botApi;

    public function __construct(\Traversable $strategies, BotApiComplete $botApi, private readonly LoggerInterface $logger)
    {
        $this->strategies = iterator_to_array($strategies);
        $this->botApi     = $botApi;

        usort($this->strategies, fn (StrategyInterface $first, StrategyInterface $second) => $first->getPriority() <= $second->getPriority());
    }

    /**
     * @throws TGException\BadArgumentException
     * @throws TGException\ResponseException
     */
    public function process(Input $input, ConfigurationInterface $configuration): void
    {
        try {
            foreach ($this->strategies as $instance) {
                if ($instance->canProcess($configuration->getStep(), $input)) {
                    try {
                        $output = $instance->process($input, $configuration);
                    } catch (SkipStrategyException $exception) {
                        unset($exception);

                        continue;
                    }

                    if (!\is_array($output)) {
                        $output = [$output];
                    }

                    foreach ($output as $item) {
                        $item->send($this->botApi, $configuration);
                    }

                    return;
                }
            }
        } catch (\Throwable $exception) {
            $this->logger->error($exception->getMessage(), [
                'file' => $exception->getFile(),
                'line' => $exception->getLine(),
            ]);

            $output = new BasicMessage('Мне что-то нехорошо от таких вопросов... Попробуйте спросить немного позже.', null);
            $method = SendMessageMethod::create($configuration->getChatId(), $output->getMessage());
            $this->botApi->sendMessage($method);

            return;
        }

        $output = new BasicMessage('Простите, я вас не понимаю, повторите ввод или нажмите /start', null);
        $method = SendMessageMethod::create($configuration->getChatId(), $output->getMessage());
        $this->botApi->sendMessage($method);
    }
}
