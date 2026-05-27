<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Repository;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Entity\ExchangeSymbol;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\ORM\Query\Expr\Join;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<ExchangeSymbol>
 *
 * @method null|ExchangeSymbol find($id, $lockMode = null, $lockVersion = null)
 * @method null|ExchangeSymbol findOneBy(array $criteria, array $orderBy = null)
 * @method ExchangeSymbol[]    findAll()
 * @method ExchangeSymbol[]    findBy(array $criteria, array $orderBy = null, $limit = null, $offset = null)
 */
class ExchangeSymbolRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, ExchangeSymbol::class);
    }

    public function add(ExchangeSymbol $entity, bool $flush = false): void
    {
        $this->getEntityManager()->persist($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function remove(ExchangeSymbol $entity, bool $flush = false): void
    {
        $this->getEntityManager()->remove($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function getAvailableSymbols(CryptoBot $cryptoBot): array
    {
        $builder = $this->createQueryBuilder('exs');

        return $builder
            ->leftJoin(
                CryptoTradeConfig::class,
                'cnf',
                Join::WITH,
                'cnf.symbol = exs.symbol AND cnf.cryptobot = :bot AND exs.exchange = :exchange'
            )
            ->andWhere($builder->expr()->isNull('cnf.id'))
            ->andWhere($builder->expr()->eq('exs.enabled', ':enabled'))
            ->andWhere($builder->expr()->eq('exs.exchange', ':exchange'))
            ->setParameter('exchange', $cryptoBot->getProvider())
            ->setParameter('bot', $cryptoBot->getId())
            ->setParameter('enabled', true)
            ->orderBy('exs.symbol', 'ASC')
            ->getQuery()
            ->getResult();
    }

    public function getSymbolCoins(): array
    {
        $sql = <<<EOL
SELECT
    DISTINCT(es.symbol) as symbol,
    SUBSTR(symbol, 1, LENGTH(symbol)-4) as coin
FROM exchange_symbol es
WHERE es.enabled = 1
GROUP BY symbol
EOL;

        return $this->getEntityManager()
            ->getConnection()
            ->executeQuery($sql)
            ->fetchAllAssociative();
    }
}
