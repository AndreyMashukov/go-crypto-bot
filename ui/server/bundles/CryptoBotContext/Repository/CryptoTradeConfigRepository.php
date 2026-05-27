<?php
namespace Bundles\CryptoBotContext\Repository;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Model\Signal;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<CryptoTradeConfig>
 */
class CryptoTradeConfigRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, CryptoTradeConfig::class);
    }

    public function getSignalSubscribers(Signal $signal): array
    {
        $builder = $this->createQueryBuilder('cfg');

        return $builder
            ->innerJoin('cfg.cryptobot', 'bot')
            ->innerJoin('bot.user', 'usr')
            ->andWhere($builder->expr()->andX(
                $builder->expr()->isNotNull('usr.signalSubscriptionExpiresAt'),
                $builder->expr()->gt('usr.signalSubscriptionExpiresAt', ':now')
            ))
            ->andWhere($builder->expr()->eq('bot.status', ':running'))
            ->andWhere($builder->expr()->eq('bot.provider', ':exchange'))
            ->andWhere($builder->expr()->eq('cfg.signalTrading', ':enabled'))
            ->andWhere($builder->expr()->eq('cfg.signalConfig.signalPeriodDays', ':days'))
            ->andWhere($builder->expr()->eq('cfg.symbol', ':symbol'))
            ->setParameter('now', new \DateTimeImmutable())
            ->setParameter('symbol', $signal->getSymbol())
            ->setParameter('exchange', $signal->getExchange())
            ->setParameter('enabled', true)
            ->setParameter('days', $signal->getPeriodDays())
            ->setParameter('running', CryptoBot::STATUS_RUNNING)
            ->getQuery()
            ->getResult();
    }

    public function getActiveConfigs(string $symbol): array
    {
        return $this->createQueryBuilder('cfg')
            ->innerJoin('cfg.cryptobot', 'bot')
            ->andWhere('bot.status = :status')
            ->andWhere('cfg.symbol = :symbol')
            ->setParameter('status', CryptoBot::STATUS_RUNNING)
            ->setParameter('symbol', $symbol)
            ->getQuery()
            ->getResult();
    }

    public function add(CryptoTradeConfig $entity, bool $flush = false): void
    {
        $this->getEntityManager()->persist($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function remove(CryptoTradeConfig $entity, bool $flush = false): void
    {
        $this->getEntityManager()->remove($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function getSentimentList(): array
    {
        $sql = <<<EOL
SELECT
    SUBSTR(c.cfg_symbol, 1, LENGTH(c.cfg_symbol)-4) as coin,
    MAX(c.cfg_score) as score,
    MAX(c.cfg_label) as label,
    JSON_ARRAYAGG(JSON_OBJECT('url', r.rsa_url, 'score', r.rsa_score, 'label', r.rsa_label)) as news
FROM crypto_trade_config c
INNER JOIN rss_article r ON r.rsa_expired = 0 AND JSON_SEARCH(r.rsa_coins, 'one', SUBSTR(c.cfg_symbol, 1, LENGTH(c.cfg_symbol)-4)) is not null
GROUP BY c.cfg_symbol
EOL;
        $result = $this->getEntityManager()
            ->getConnection()
            ->executeQuery($sql)
            ->fetchAllAssociative();

        foreach ($result as $key => $item) {
            $news = json_decode((string) $item['news'], true);
            $uniq = [];
            foreach ($news as $article) {
                $uniq[$article['url']] = $article;
            }

            $result[$key]['news'] = array_values($uniq);
        }

        return $result;
    }
}
