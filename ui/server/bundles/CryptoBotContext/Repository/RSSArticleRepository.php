<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Repository;

use Bundles\CryptoBotContext\Entity\RSSArticle;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<RSSArticle>
 *
 * @method null|RSSArticle find($id, $lockMode = null, $lockVersion = null)
 * @method null|RSSArticle findOneBy(array $criteria, array $orderBy = null)
 * @method RSSArticle[]    findAll()
 * @method RSSArticle[]    findBy(array $criteria, array $orderBy = null, $limit = null, $offset = null)
 */
class RSSArticleRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, RSSArticle::class);
    }

    public function add(RSSArticle $entity, bool $flush = false): void
    {
        $this->getEntityManager()->persist($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function remove(RSSArticle $entity, bool $flush = false): void
    {
        $this->getEntityManager()->remove($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function expireArticles(): void
    {
        $sql = <<<EOL
UPDATE rss_article SET rsa_expired = 1 WHERE rsa_expires_at <= now()
EOL;

        $this->getEntityManager()
            ->getConnection()
            ->executeQuery($sql);
    }

    public function getResults(string $coin): array
    {
        $results = [
            'BEARISH' => [0, 0.00],
            'BULLISH' => [0, 0.00],
            'NEUTRAL' => [0, 0.00],
        ];

        $sql = <<<EOL
SELECT
    r.rsa_id,
    r.rsa_label as label,
    r.rsa_score as score
FROM rss_article r
WHERE JSON_SEARCH(r.rsa_coins, 'one', ?) is not null AND r.rsa_expired = 0 AND r.rsa_expires_at > now()
EOL;
        $data = $this->getEntityManager()
            ->getConnection()
            ->executeQuery($sql, [$coin])
            ->fetchAllAssociative();

        if (!$data) {
            return [];
        }

        foreach ($data as $item) {
            ++$results[$item['label']][0];
            $results[$item['label']][1] += $item['score'];
        }

        return $results;
    }
}
