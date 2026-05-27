<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Command;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Repository\CryptoBotRepository;
use Bundles\CryptoBotContext\Service\CryptoBotService;
use Doctrine\DBAL\LockMode;
use Doctrine\ORM\EntityManagerInterface;
use Psr\Log\LoggerInterface;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;

class BotPositionPriceCommand extends Command
{
    protected static $defaultName = 'bot:position:price';

    protected static $defaultDescription = 'Update position price information';

    private CryptoBotRepository $repository;

    private CryptoBotService $cryptoBotService;

    private LoggerInterface $logger;

    private EntityManagerInterface $entityManager;

    public function __construct(
        CryptoBotRepository $repository,
        CryptoBotService $cryptoBotService,
        LoggerInterface $logger,
        EntityManagerInterface $entityManager
    ) {
        parent::__construct(self::$defaultName);

        $this->repository       = $repository;
        $this->cryptoBotService = $cryptoBotService;
        $this->logger           = $logger;
        $this->entityManager    = $entityManager;
    }

    protected function configure(): void
    {
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        $cryptoBots = $this->repository->findBy([
            'status' => CryptoBot::STATUS_RUNNING,
        ]);

        $now = new \DateTimeImmutable();

        foreach ($cryptoBots as $cryptoBot) {
            try {
                $positions = $this->cryptoBotService->getPositionsList($cryptoBot);
            } catch (\Throwable $exception) {
                $this->logger->error($exception->getMessage(), [
                    'file' => $exception->getFile(),
                    'line' => $exception->getLine(),
                ]);
                continue;
            }

            $configMap = [];
            foreach ($cryptoBot->getCryptoTradeConfigs() as $config) {
                $configMap[$config->getSymbol()] = $config;
            }

            foreach ($positions as $position) {
                if (!isset($position['kLine']['c']) || !isset($position['order'])) {
                    unset($configMap[$position['symbol']]);
                    continue;
                }

                $config = $configMap[$position['symbol']] ?? null;
                if (!$config instanceof CryptoTradeConfig) {
                    continue;
                }

                $this->entityManager->wrapInTransaction(function () use ($config, $now, $position) {
                    $this->entityManager->refresh($config);
                    $this->entityManager->lock($config, LockMode::PESSIMISTIC_WRITE);

                    $lastUpdate = $config->getPositionUpdatedAt();
                    if (!$lastUpdate instanceof \DateTimeImmutable || $lastUpdate->format('Y-m-d') !== $now->format('Y-m-d')) {
                        $config->setPriceTodayFirst(null);
                    }

                    if (!$config->getPriceTodayFirst()) {
                        $config->setPriceTodayFirst($position['kLine']['c']);
                    }

                    $config
                        ->setAveragePrice($position['order']['price'])
                        ->setQuantity($position['order']['executedQuantity'])
                        ->setPriceTodayLast($position['kLine']['c'])
                        ->setPositionUpdatedAt($now)
                    ;
                });

                unset($configMap[$position['symbol']]);
            }

            foreach ($configMap as $configWithoutPosition) {
                $this->entityManager->wrapInTransaction(function () use ($configWithoutPosition) {
                    $this->entityManager->refresh($configWithoutPosition);
                    $this->entityManager->lock($configWithoutPosition, LockMode::PESSIMISTIC_WRITE);

                    $configWithoutPosition
                        ->setAveragePrice(null)
                        ->setPriceTodayFirst(null)
                        ->setPriceTodayLast(null)
                        ->setQuantity(null)
                        ->setPositionUpdatedAt(null)
                    ;
                });
            }
        }

        return Command::SUCCESS;
    }
}
