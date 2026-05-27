<?php
namespace Bundles\TgBotContext\Model\Traits;

trait SendTrait
{
    /**
     * @throws \Throwable
     */
    public function handleSend(callable $callback): void
    {
        try {
            $callback();
        } catch (\Throwable $throwable) {
            if (preg_match('/Too Many Requests: retry after (?P<seconds>\d+)/ui', $throwable->getMessage(), $matches)) {
                $seconds = $matches['seconds'] ?? 0;
                sleep($seconds);
                $this->handleSend($callback);

                return;
            }

            throw $throwable;
        }
    }
}
