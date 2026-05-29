<?php

declare(strict_types=1);

namespace App\Service\Monitoring\PrometheusQuery;

use Psr\Log\LoggerInterface;
use Symfony\Contracts\HttpClient\Exception\ExceptionInterface;
use Symfony\Contracts\HttpClient\HttpClientInterface;

final readonly class PrometheusQueryService implements PrometheusQueryServiceInterface
{
    public function __construct(
        private HttpClientInterface $httpClient,
        private LoggerInterface $logger,
        private string $prometheusUrl,
    ) {}

    public function queryRange(
        string $promQL,
        \DateTimeImmutable $from,
        \DateTimeImmutable $to,
        string $step,
    ): array {
        try {
            $response = $this->httpClient->request(
                'GET',
                \sprintf('%s/api/v1/query_range', $this->prometheusUrl),
                [
                    'query' => [
                        'query' => $promQL,
                        'start' => $from->format(\DATE_RFC3339),
                        'end'   => $to->format(\DATE_RFC3339),
                        'step'  => $step,
                    ],
                    'timeout'      => 5,
                    'max_duration' => 6,
                ],
            );

            $data = $response->toArray(false);

            if ('success' !== ($data['status'] ?? null)) {
                $this->logger->error('Prometheus range query returned non-success', [
                    'promQL' => $promQL,
                    'status' => $data['status'] ?? 'unknown',
                    'error'  => $data['error'] ?? null,
                ]);

                return ['series' => []];
            }

            return $this->parse($data);
        } catch (ExceptionInterface | \JsonException $exception) {
            $this->logger->error('Failed to query Prometheus', [
                'promQL'    => $promQL,
                'exception' => $exception->getMessage(),
            ]);

            return ['series' => []];
        }
    }

    /**
     * @param array<string, mixed> $data
     *
     * @return array{series: list<array{label: string, points: list<array{0: int, 1: string}>}>}
     */
    private function parse(array $data): array
    {
        $result = $data['data']['result'] ?? [];
        if (!\is_array($result)) {
            return ['series' => []];
        }

        $series = [];
        foreach ($result as $seriesItem) {
            $labels = $seriesItem['metric'] ?? [];
            $values = $seriesItem['values'] ?? [];

            if (!\is_array($values)) {
                continue;
            }

            $points = [];
            foreach ($values as $value) {
                if (!\is_array($value) || 2 !== \count($value)) {
                    continue;
                }
                $points[] = [(int) $value[0], (string) $value[1]];
            }

            $series[] = [
                'label'  => $this->labelString($labels),
                'points' => $points,
            ];
        }

        return ['series' => $series];
    }

    /**
     * @param array<string, mixed> $labels
     */
    private function labelString(array $labels): string
    {
        $name = (string) ($labels['__name__'] ?? '');
        $parts = [];
        foreach ($labels as $key => $value) {
            if ('__name__' === $key) {
                continue;
            }
            $parts[] = \sprintf('%s="%s"', $key, (string) $value);
        }

        if ([] === $parts) {
            return $name;
        }

        return $name.'{'.implode(',', $parts).'}';
    }
}
