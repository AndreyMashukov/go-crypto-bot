<?php

declare(strict_types=1);

namespace App\Service\Monitoring\PrometheusQuery;

interface PrometheusQueryServiceInterface
{
    /**
     * Proxies a Prometheus range query and returns the architecture-doc
     * §8.3 JSON envelope:
     *
     *     {
     *         "series": [
     *             {
     *                 "label": "<metric or labelset>",
     *                 "points": [[<unix-ts>, "<value>"], ...]
     *             },
     *             ...
     *         ]
     *     }
     *
     * Errors are absorbed and surface as an empty {"series": []} so the
     * FE chart never blocks on transient Prometheus hiccups.
     *
     * @return array{series: list<array{label: string, points: list<array{0: int, 1: string}>}>}
     */
    public function queryRange(
        string $promQL,
        \DateTimeImmutable $from,
        \DateTimeImmutable $to,
        string $step,
    ): array;
}
