<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Command;

use Bundles\CryptoBotContext\Repository\CryptoBotRepository;
use Bundles\CryptoBotContext\Service\TradeSyncManager;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;
use Symfony\Component\Console\Style\SymfonyStyle;

class BotTradesSyncCommand extends Command
{
    protected static $defaultName = 'bot:trades:sync';

    protected static $defaultDescription = 'Sync bots trade list';

    private CryptoBotRepository $repository;

    private TradeSyncManager $tradeSyncManager;

    public function __construct(
        CryptoBotRepository $repository,
        TradeSyncManager $tradeSyncManager
    ) {
        parent::__construct(self::$defaultName);

        $this->repository       = $repository;
        $this->tradeSyncManager = $tradeSyncManager;
    }

    protected function configure(): void
    {
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        $io = new SymfonyStyle($input, $output);

        foreach ($this->repository->findAll() as $cryptoBot) {
            if (!$cryptoBot->getServer()) {
                continue;
            }

            $this->tradeSyncManager->syncTrades($cryptoBot, function (\Throwable $throwable) use ($io) {
                $io->error($throwable);
            }, true);
        }

        return Command::SUCCESS;
    }
}
