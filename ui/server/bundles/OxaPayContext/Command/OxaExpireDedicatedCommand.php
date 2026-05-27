<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\OxaPayContext\Command;

use Bundles\CryptoBotContext\Repository\CryptoBotRepository;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;
use Symfony\Component\Console\Style\SymfonyStyle;

class OxaExpireDedicatedCommand extends Command
{
    protected static $defaultName = 'oxa:expire:dedicated';

    protected static $defaultDescription = 'Unset expired dedicated servers from bots';

    private CryptoBotRepository $cryptoBotRepository;

    public function __construct(CryptoBotRepository $cryptoBotRepository)
    {
        parent::__construct(self::$defaultName);

        $this->cryptoBotRepository = $cryptoBotRepository;
    }

    protected function configure(): void
    {
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        $io      = new SymfonyStyle($input, $output);
        $expired = 0;

        foreach ($this->cryptoBotRepository->getBotsWithExpiredDedicatedServer() as $cryptobot) {
            $cryptobot->setDedicated(null);
            $this->cryptoBotRepository->add($cryptobot, true);
            ++$expired;
            // todo: notify user about expired dedicated server
        }

        $io->success("Expired: {$expired} dedicated server for bots.");

        return Command::SUCCESS;
    }
}
