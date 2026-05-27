<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Command;

use Bundles\CryptoBotContext\Repository\CryptoBotRepository;
use Bundles\CryptoBotContext\Service\DeployDomain;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;
use Symfony\Component\Console\Style\SymfonyStyle;

class BotStopAllCommand extends Command
{
    protected static $defaultName = 'bot:stop:all';

    protected static $defaultDescription = 'Stop all bots';

    private CryptoBotRepository $repository;

    private DeployDomain $deployDomain;

    public function __construct(
        CryptoBotRepository $repository,
        DeployDomain $deployDomain
    ) {
        parent::__construct(self::$defaultName);

        $this->repository   = $repository;
        $this->deployDomain = $deployDomain;
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        $io = new SymfonyStyle($input, $output);

        foreach ($this->repository->findAll() as $cryptoBot) {
            try {
                if ($cryptoBot->getServer()) {
                    $this->deployDomain->doStop(
                        $cryptoBot,
                        $cryptoBot->getServer(),
                        '',
                        true
                    );
                }
                $io->note("Bot {$cryptoBot->getId()} has stopped");
            } catch (\Exception $exception) {
                $io->error($exception);
            }
        }

        return Command::SUCCESS;
    }
}
