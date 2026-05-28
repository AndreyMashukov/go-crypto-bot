<?php
declare(strict_types=1);

namespace Bundles\CryptoBotContext\Command;

use Bundles\CryptoBotContext\Service\Worker\SentimentBenchmarkWorker;
use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Input\InputOption;
use Symfony\Component\Console\Output\OutputInterface;
use Symfony\Component\DependencyInjection\Attribute\Autowire;

#[AsCommand(name: 'parallax:benchmark', description: 'Fan-out 23 symbols serially vs through php-parallax and print the speed-up.')]
final class ParallaxBenchmarkCommand extends Command
{
    private const TRADE_LIMIT_SYMBOLS = [
        'NEOUSDT', 'PERPUSDT', 'ETHUSDT', 'SOLUSDT', 'BTCUSDT', 'LTCUSDT',
        'XRPUSDT', 'BNBUSDT', 'TRXUSDT', 'AVAXUSDT', 'ADAUSDT', 'DOGEUSDT',
        'BCHUSDT', 'LINKUSDT', 'MATICUSDT', 'DOTUSDT', 'UNIUSDT', 'ETCUSDT',
        'XLMUSDT', 'ATOMUSDT', 'NEARUSDT', 'ZECUSDT', 'SHIBUSDT',
    ];

    public function __construct(
        #[Autowire(param: 'kernel.project_dir')]
        private string $projectDir,
    ) {
        parent::__construct();
    }

    protected function configure(): void
    {
        $this
            ->addOption('per-task-ms', null, InputOption::VALUE_OPTIONAL, 'Simulated per-symbol work in milliseconds.', '120')
            ->addOption('cpu-iters', null, InputOption::VALUE_OPTIONAL, 'Per-symbol CPU work iterations.', '2000000')
        ;
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        if (!class_exists(\WaitGroup::class)) {
            $output->writeln('<error>parallax extension is not loaded — abort.</error>');

            return Command::FAILURE;
        }

        $perTaskMs = (int) $input->getOption('per-task-ms');
        $cpuIters  = (int) $input->getOption('cpu-iters');
        $symbols   = self::TRADE_LIMIT_SYMBOLS;

        $output->writeln(sprintf(
            '<info>parallax benchmark — %d symbols, ~%d ms simulated I/O + %d CPU iters per task</info>',
            \count($symbols),
            $perTaskMs,
            $cpuIters,
        ));

        $serialStart = hrtime(true);
        foreach ($symbols as $symbol) {
            SentimentBenchmarkWorker::work($symbol, $perTaskMs, $cpuIters);
        }
        $serialMs = (hrtime(true) - $serialStart) / 1_000_000.0;

        $bootstrap = $this->projectDir . '/config/parallax_bootstrap.php';

        $parallelStart = hrtime(true);
        $wg            = new \WaitGroup($bootstrap);
        foreach ($symbols as $symbol) {
            $wg->go(SentimentBenchmarkWorker::work(...), [$symbol, $perTaskMs, $cpuIters]);
        }
        $results    = $wg->wait();
        $parallelMs = (hrtime(true) - $parallelStart) / 1_000_000.0;

        $failed = 0;
        foreach ($results as $slot) {
            if (!$slot->ok) {
                $failed++;
                $output->writeln(sprintf(
                    '<comment>slot failed: %s — %s</comment>',
                    $slot->error?->class ?? '?',
                    $slot->error?->message ?? '?',
                ));
            }
        }

        $speedup = $serialMs / max(0.001, $parallelMs);

        $output->writeln('');
        $output->writeln(sprintf('  serial   = %8.1f ms', $serialMs));
        $output->writeln(sprintf('  parallel = %8.1f ms', $parallelMs));
        $output->writeln(sprintf('  speed-up = %8.2fx (%d failed slot(s))', $speedup, $failed));

        return $failed === 0 ? Command::SUCCESS : Command::FAILURE;
    }
}
