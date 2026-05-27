<?php
namespace Bundles\CryptoBotContext\Repository;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<CryptoBot>
 */
class CryptoBotRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, CryptoBot::class);
    }

    public function add(CryptoBot $entity, bool $flush = false): void
    {
        $this->getEntityManager()->persist($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function remove(CryptoBot $entity, bool $flush = false): void
    {
        $this->getEntityManager()->remove($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    /**
     * @return array|CryptoBot[]
     */
    public function getBotsWithExpiredDedicatedServer(): array
    {
        $builder = $this->createQueryBuilder('ctb');

        return $builder->andWhere($builder->expr()->isNotNull('ctb.dedicated'))
            ->andWhere($builder->expr()->lte('ctb.dedicatedServerExpiresAt', ':now'))
            ->setParameter('now', new \DateTimeImmutable('now'))
            ->getQuery()
            ->getResult();
    }
}
