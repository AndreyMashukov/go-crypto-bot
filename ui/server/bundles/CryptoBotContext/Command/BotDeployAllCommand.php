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

class BotDeployAllCommand extends Command
{
    protected static $defaultName = 'bot:deploy:all';

    protected static $defaultDescription = 'Deploy all bots (paid only)';

    private CryptoBotRepository $repository;

    private DeployDomain $deployDomain;

    public function __construct(CryptoBotRepository $repository, DeployDomain $deployDomain)
    {
        parent::__construct(self::$defaultName);

        $this->repository    = $repository;
        $this->deployDomain  = $deployDomain;
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        $io = new SymfonyStyle($input, $output);

        foreach ($this->repository->findAll() as $cryptoBot) {
            if (!$cryptoBot->isPaid()) {
                $io->warning("Bot {$cryptoBot->getId()} is not paid.");
                continue;
            }

            try {
                $this->deployDomain->doDeploy($cryptoBot);
                $io->note("Bot {$cryptoBot->getId()} has deployed");
            } catch (\Exception $exception) {
                $io->error($exception->getMessage());
            }
        }

        return Command::SUCCESS;
    }
}
