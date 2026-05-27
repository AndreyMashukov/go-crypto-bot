<?php
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

    public function __construct(private readonly CryptoBotRepository $repository, private readonly DeployDomain $deployDomain)
    {
        parent::__construct(self::$defaultName);
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
