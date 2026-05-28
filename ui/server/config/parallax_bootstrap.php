<?php
declare(strict_types=1);

/**
 * Bootstrap loaded inside each parallax worker thread via
 * `new WaitGroup(__DIR__ . '/parallax_bootstrap.php')`.
 *
 * The file pulls in the Composer autoloader so user-defined classes
 * (App\, Bundles\, vendor packages) resolve inside the fresh worker
 * request lifecycle. It deliberately stops there: instantiating the
 * Symfony container per worker would defeat the parallax win.
 */

require dirname(__DIR__) . '/vendor/autoload.php';
