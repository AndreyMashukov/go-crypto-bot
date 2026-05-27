<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Repository;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\Trade;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<Trade>
 */
class TradeRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, Trade::class);
    }

    public function add(Trade $entity, bool $flush = false): void
    {
        $this->getEntityManager()->persist($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function remove(Trade $entity, bool $flush = false): void
    {
        $this->getEntityManager()->remove($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function getProfitByPeriod(CryptoBot $cryptobot, string $period): array
    {
        $query = $this->getSqlProfitPeriod($period);

        if ($query === '' || $query === '0') {
            return [];
        }

        try {
            return $this->getEntityManager()
                ->getConnection()
                ->executeQuery($query, [$cryptobot->getId()])
                ->fetchAllAssociative();
        } catch (\Throwable $throwable) {
            unset($throwable);

            return [];
        }
    }

    public function getBestMonthSymbols(): array
    {
        $sql = <<<EOL
SELECT
    AVG(trade.trd_buy_price) as avgBuyPrice,
    AVG(trade.trd_sell_price) as avgSellPrice,
    ROUND(AVG(trade.trd_sell_price) * 100 / AVG(trade.trd_buy_price) - 100, 2) as percent,
    ROUND(AVG(positionTimeHours), 0) as avgPositionTimeHours,
    ROUND(SUM(trade.trd_profit), 2) as totalProfit,
    COUNT(trade.trd_id) as trades,
    ROUND(SUM(trade.trd_profit) / COUNT(trd_id), 2) as profitPerTrade,
    trade.trd_symbol as symbol
FROM (
         SELECT
             t.trd_buy_price,
             t.trd_sell_price,
             (UNIX_TIMESTAMP(t.trd_sell_date) - UNIX_TIMESTAMP(t.trd_buy_date)) / 3600 as positionTimeHours,
             t.trd_profit,
             t.trd_id,
             t.trd_symbol
         FROM trade t
         INNER JOIN crypto_bot b ON b.ctb_id = t.trd_bot AND b.ctb_test = 0
         WHERE t.trd_sell_date >= current_date() - interval 30 DAY AND (UNIX_TIMESTAMP(t.trd_sell_date) - UNIX_TIMESTAMP(t.trd_buy_date)) / 3600 < 72
     ) as trade
GROUP BY trade.trd_symbol
HAVING trades > 1
ORDER BY totalProfit DESC
EOL;

        try {
            $map    = [];
            $result = $this->getEntityManager()
                ->getConnection()
                ->executeQuery($sql)
                ->fetchAllAssociative();

            foreach ($result as $index => $item) {
                $map[$item['symbol']] = [
                    'rating'               => $index + 1,
                    'totalProfit'          => $item['totalProfit'],
                    'avgBuyPrice'          => $item['avgBuyPrice'],
                    'avgSellPrice'         => $item['avgSellPrice'],
                    'percent'              => $item['percent'],
                    'profitPerTrade'       => $item['profitPerTrade'],
                    'trades'               => $item['trades'],
                    'avgPositionTimeHours' => $item['avgPositionTimeHours'],
                ];
            }

            return $map;
        } catch (\Throwable $throwable) {
            unset($throwable);

            return [];
        }
    }

    private function getSqlProfitPeriod(string $period): string
    {
        return match ($period) {
            'month' => <<<EOL
SELECT
    DATE_FORMAT(t.trd_sell_date, '%Y-%m') AS title,
    SUM(t.trd_profit) AS profit,
    MAX(t.trd_sell_date) as maxDate,
    COUNT(t.trd_id) as trades,
    SUM(t.trd_buy_qty * t.trd_buy_price) as tradeVolume,
    AVG(t.trd_percent) as avgPercent
FROM trade t
WHERE t.trd_bot = ?
GROUP BY Title
ORDER BY MaxDate DESC
EOL,
            'week' => <<<EOL
SELECT
    CONCAT(DATE_FORMAT(t.trd_sell_date, '%Y-%m-'), WEEK(t.trd_sell_date)) AS title,
    SUM(t.trd_profit) AS profit,
    MAX(t.trd_sell_date) as maxDate,
    COUNT(t.trd_id) as trades,
    SUM(t.trd_buy_qty * t.trd_buy_price) as tradeVolume,
    AVG(t.trd_percent) as avgPercent
FROM trade t
WHERE t.trd_bot = ?
GROUP BY Title
ORDER BY MaxDate DESC
EOL,
            'day' => <<<EOL
SELECT
    DATE_FORMAT(t.trd_sell_date, '%Y-%m-%d') AS title,
    SUM(t.trd_profit) AS profit,
    MAX(t.trd_sell_date) as maxDate,
    COUNT(t.trd_id) as trades,
    SUM(t.trd_buy_qty * t.trd_buy_price) as tradeVolume,
    AVG(t.trd_percent) as avgPercent
FROM trade t
WHERE t.trd_bot = ?
GROUP BY Title
ORDER BY MaxDate DESC
EOL,
            default => '',
        };
    }
}
