<?php
namespace Bundles\CryptoBotContext\Command;

use Bundles\CryptoBotContext\Entity\RSSArticle;
use Bundles\CryptoBotContext\Repository\CryptoTradeConfigRepository;
use Bundles\CryptoBotContext\Repository\ExchangeSymbolRepository;
use Bundles\CryptoBotContext\Repository\RSSArticleRepository;
use Bundles\CryptoBotContext\Service\RSSParser;
use Bundles\CryptoBotContext\Service\SentimentService;
use Bundles\CryptoBotContext\Service\Traits\SentimentAverageScoreTrait;
use Bundles\TgBotContext\Service\AlertService;
use Doctrine\ORM\EntityManagerInterface;
use Psr\Log\LoggerInterface;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;

class BotSentimentProcessCommand extends Command
{
    use SentimentAverageScoreTrait;

    protected static $defaultName = 'bot:sentiment:process';

    protected static $defaultDescription = 'RSS feed sentiment analyser';

    public function __construct(
        private RSSParser $rssParser,
        private SentimentService $sentimentService,
        private RSSArticleRepository $repository,
        private ExchangeSymbolRepository $symbolRepository,
        private CryptoTradeConfigRepository $configRepository,
        private EntityManagerInterface $entityManager,
        private LoggerInterface $logger,
        private AlertService $alertService
    ) {
        parent::__construct(self::$defaultName);
    }

    protected function configure(): void
    {
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        foreach ($this->rssParser->iterateArticles() as $article) {
            try {
                $sentimentResult  = $this->sentimentService->sentimentAnalysis($article);
            } catch (\Throwable $throwable) {
                $this->logger->error($throwable->getMessage(), [
                    'line' => $throwable->getLine(),
                    'file' => $throwable->getFile(),
                ]);
                continue;
            }

            $rssArticleEntity = new RSSArticle();
            $rssArticleEntity
                ->setCoins($sentimentResult->getCoins())
                ->setLabel($sentimentResult->getLabel())
                ->setScore($sentimentResult->getScore())
                ->setUrl($sentimentResult->getArticle()->getUrl())
                ->setUrlCrc32(crc32($sentimentResult->getArticle()->getUrl()))
                ->setExpiresAt(new \DateTimeImmutable('+1 day'))
            ;

            $this->repository->add($rssArticleEntity, true);
        }

        $this->repository->expireArticles();

        $changed = [];

        foreach ($this->symbolRepository->getSymbolCoins() as $symbolCoin) {
            $configs = $this->configRepository->getActiveConfigs($symbolCoin['symbol']);
            $results = $this->repository->getResults($symbolCoin['coin']);

            if ($results !== []) {
                [$label, $score] = $this->processAverage($this->calculateAverage($results));

                foreach ($configs as $config) {
                    if ($label !== $config->getLabel() && $config->getCryptobot()->isMaster()) {
                        $changed[$config->getSymbol()] = [
                            $config->getLabel() ?: 'null',
                            $config->getScore() ?: '0.00',
                            $label,
                            $score,
                        ];
                    }

                    $config
                        ->setScore($score)
                        ->setLabel($label)
                    ;
                }
            } else {
                foreach ($configs as $config) {
                    if ($config->getLabel() && $config->getCryptobot()->isMaster()) {
                        $changed[$config->getSymbol()] = [
                            $config->getLabel(),
                            $config->getScore(),
                            'null',
                            'null',
                        ];
                    }

                    $config
                        ->setScore(null)
                        ->setLabel(null)
                    ;
                }
            }
            $this->entityManager->flush();
            $this->entityManager->clear();
        }

        foreach ($changed as $symbol => $list) {
            [$oldLabel, $oldScore, $newLabel, $newScore] = $list;
            $this->alertService->alert("BotSentimentProcessCommand: [{$symbol}] sentiment changed {$oldLabel}[{$oldScore}] -> {$newLabel}[{$newScore}]");
        }

        return Command::SUCCESS;
    }
}
