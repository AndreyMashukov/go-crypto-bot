<?php
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

    public function __construct(
        private readonly CryptoBotRepository $repository,
        private readonly DeployDomain $deployDomain
    ) {
        parent::__construct(self::$defaultName);
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
