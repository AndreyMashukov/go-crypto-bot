<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

use Symfony\Component\Dotenv\Dotenv;

require \dirname(__DIR__) . '/vendor/autoload.php';

if (file_exists(\dirname(__DIR__) . '/config/bootstrap.php')) {
    require \dirname(__DIR__) . '/config/bootstrap.php';
} elseif (method_exists(Dotenv::class, 'bootEnv')) {
    (new Dotenv())->bootEnv(\dirname(__DIR__) . '/.env');
}
