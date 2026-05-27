<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Repository;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\Server;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<Server>
 *
 * @method null|Server find($id, $lockMode = null, $lockVersion = null)
 * @method null|Server findOneBy(array $criteria, array $orderBy = null)
 * @method Server[]    findAll()
 * @method Server[]    findBy(array $criteria, array $orderBy = null, $limit = null, $offset = null)
 */
class ServerRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, Server::class);
    }

    public function add(Server $entity, bool $flush = false): void
    {
        $this->getEntityManager()->persist($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function remove(Server $entity, bool $flush = false): void
    {
        $this->getEntityManager()->remove($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function getAvailableServer(CryptoBot $cryptoBot): ?Server
    {
        if ($cryptoBot->getServer()) {
            return $cryptoBot->getServer();
        }

        if ($cryptoBot->getDedicated()) {
            return $cryptoBot->getDedicated();
        }

        $stats = $this->getServerStats();

        if (0 === \count($stats)) {
            return null;
        }

        $available = $stats[0]['available'];
        if ($available <= 0) {
            return null;
        }

        return $this->find($stats[0]['id']);
    }

    public function getServerStats(): array
    {
        $sql = <<<EOL
SELECT
    s.srv_id as id,
    s.srv_ip as ip,
    s.srv_slots as slots,
    SUM(IF(cb.ctb_id is NULL, 0, 1)) as booked,
    (s.srv_slots - SUM(IF(cb.ctb_id is NULL, 0, 1))) as available
FROM server s
LEFT JOIN
    crypto_bot cb ON cb.ctb_server = s.srv_id OR cb.ctb_dedicated_server = s.srv_id
GROUP BY s.srv_id
ORDER BY available DESC
EOL;

        return $this->getEntityManager()
            ->getConnection()
            ->executeQuery($sql)
            ->fetchAllAssociative();
    }
}
