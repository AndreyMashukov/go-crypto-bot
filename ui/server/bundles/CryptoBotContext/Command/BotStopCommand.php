<?php
namespace Bundles\CryptoBotContext\Command;

use Bundles\CryptoBotContext\Entity\Server;
use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Repository\CryptoBotRepository;
use Bundles\CryptoBotContext\Service\DeployDomain;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Input\InputOption;
use Symfony\Component\Console\Output\OutputInterface;
use Symfony\Component\Console\Style\SymfonyStyle;

class BotStopCommand extends Command
{
    protected static $defaultName = 'bot:stop';

    protected static $defaultDescription = 'Stop bot';

    public function __construct(
        private readonly CryptoBotRepository $repository,
        private readonly DeployDomain $deployDomain
    ) {
        parent::__construct(self::$defaultName);
    }

    protected function configure()
    {
        $this->addOption('uuid', '-u', InputOption::VALUE_REQUIRED, 'Bot UUID');
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        $io   = new SymfonyStyle($input, $output);
        $uuid = $input->getOption('uuid');

        $cryptoBot = $this->repository->findOneBy([
            'uuid' => $uuid,
        ]);

        if (!$cryptoBot instanceof CryptoBot) {
            $io->error("Bot with uuid = {$uuid} is not found.");

            return self::FAILURE;
        }

        try {
            if ($cryptoBot->getServer() instanceof Server) {
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

            return self::FAILURE;
        }

        return self::SUCCESS;
    }
}
