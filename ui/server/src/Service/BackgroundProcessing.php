<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Service;

use Psr\Log\LoggerInterface;

class BackgroundProcessing
{
    private array $tasks = [];

    private LoggerInterface $logger;

    public function __construct(LoggerInterface $logger)
    {
        $this->logger = $logger;
    }

    public function addTask(callable $task, string $uniqKey = null): void
    {
        if ($uniqKey) {
            $this->tasks[$uniqKey] = $task;
        } else {
            $this->tasks[] = $task;
        }
    }

    public function runTasks(): void
    {
        foreach ($this->tasks as $key => $task) {
            try {
                $task();
            } catch (\Throwable $exception) {
                $this->logger->error($exception->getMessage(), [
                    'file' => $exception->getFile(),
                    'line' => $exception->getLine(),
                ]);
            }

            unset($this->tasks[$key]);
        }
    }
}
