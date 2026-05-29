<?php

declare(strict_types=1);

namespace App\Controller\V1;

use App\Service\Monitoring\PrometheusQuery\PrometheusQueryServiceInterface;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;
use Symfony\Component\Routing\Attribute\Route;

/**
 * Phase F endpoint that proxies Prometheus range queries to the FE.
 * Returns the architecture-doc §8.3 JSON envelope verbatim:
 *
 *     {"series": [{"label": "...", "points": [[<unix-ts>, "<value>"], ...]}]}
 *
 * The FE renders this directly into Chart.js without further reshaping.
 */
final class PrometheusController extends AbstractController
{
    public function __construct(
        private readonly PrometheusQueryServiceInterface $prometheus,
    ) {}

    #[Route(path: '/api/prometheus/range', name: 'api_prometheus_range', methods: ['GET'])]
    public function range(Request $request): JsonResponse
    {
        $promQL = (string) $request->query->get('query', '');
        if ('' === $promQL) {
            throw new BadRequestHttpException('query parameter is required');
        }

        $step = (string) $request->query->get('step', '15s');

        $from = $this->parseTime((string) $request->query->get('start', ''), '-30 minutes');
        $to   = $this->parseTime((string) $request->query->get('end', ''), 'now');

        if ($to <= $from) {
            throw new BadRequestHttpException('end must be strictly greater than start');
        }

        return new JsonResponse($this->prometheus->queryRange($promQL, $from, $to, $step));
    }

    private function parseTime(string $raw, string $default): \DateTimeImmutable
    {
        $value = '' === $raw ? $default : $raw;

        // Accept both unix seconds and RFC3339-style strings.
        if (ctype_digit($value)) {
            return (new \DateTimeImmutable())->setTimestamp((int) $value)->setTimezone(new \DateTimeZone('UTC'));
        }

        try {
            return new \DateTimeImmutable($value, new \DateTimeZone('UTC'));
        } catch (\Exception $exception) {
            throw new BadRequestHttpException(\sprintf('invalid time value %s', $raw), $exception);
        }
    }
}
