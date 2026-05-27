<?php

declare(strict_types=1);

use Amashukov\RectorRules\NoArrayAssertContainsInTestsRector;
use Amashukov\RectorRules\NoAssertCallInSrcRector;
use Amashukov\RectorRules\NoAssertInsideIfInFunctionalTestsRector;
use Amashukov\RectorRules\NoCommentsOutsideInterfaceMethodDocBlockRector;
use Amashukov\RectorRules\NoDirectDbMutationInFunctionalTestsRector;
use Amashukov\RectorRules\NoDirectDispatchInFunctionalTestsRector;
use Amashukov\RectorRules\NoEnvironmentCheckInSrcRector;
use Amashukov\RectorRules\NoExistenceOnlyAssertionsInTestsRector;
use Amashukov\RectorRules\NoNullCoalesceNewFallbackRector;
use Amashukov\RectorRules\NoPhpstanIgnoreRector;
use Amashukov\RectorRules\NoSuperglobalAccessRector;
use Amashukov\RectorRules\NoTypeOnlyAssertionsInTestsRector;
use Amashukov\RectorRules\RequirePsrClockInterfaceRector;
use Amashukov\RectorRules\Yaml\YamlCommentStripper;
use Amashukov\RectorRules\Yaml\YamlCommentStripperInterface;
use Amashukov\RectorRules\Yaml\YamlNoCommentsChecker;
use Amashukov\RectorRules\Yaml\YamlNoCommentsCheckerInterface;
use Amashukov\RectorRules\Yaml\YamlNoCommentsRector;
use Rector\Config\RectorConfig;
use Rector\Doctrine\Set\DoctrineSetList;
use Rector\Symfony\Set\SymfonySetList;

return RectorConfig::configure()
    ->registerService(YamlNoCommentsChecker::class, YamlNoCommentsCheckerInterface::class)
    ->registerService(YamlCommentStripper::class, YamlCommentStripperInterface::class)
    ->withParallel(600, 16)
    ->withPaths([
        __DIR__ . '/src',
        __DIR__ . '/bundles',
        __DIR__ . '/tests',
    ])
    ->withImportNames(
        importNames:         true,
        importDocBlockNames: true,
        importShortClasses:  false,
        removeUnusedImports: true,
    )
    ->withSkip([
        __DIR__ . '/var',
        __DIR__ . '/vendor',
        __DIR__ . '/migrations',
        NoCommentsOutsideInterfaceMethodDocBlockRector::class => [
            __DIR__ . '/migrations',
        ],
    ])
    ->withPhpSets()
    ->withComposerBased(
        twig:     true,
        doctrine: true,
        phpunit:  true,
        symfony:  true,
    )
    ->withPreparedSets(
        deadCode:             true,
        codeQuality:          true,
        doctrineCodeQuality:  true,
        symfonyCodeQuality:   true,
        symfonyConfigs:       true,
    )
    ->withSets([
        DoctrineSetList::ANNOTATIONS_TO_ATTRIBUTES,
        SymfonySetList::ANNOTATIONS_TO_ATTRIBUTES,
    ])
    ->withRules([
        NoCommentsOutsideInterfaceMethodDocBlockRector::class,
        NoPhpstanIgnoreRector::class,
        NoSuperglobalAccessRector::class,
        NoEnvironmentCheckInSrcRector::class,
        NoAssertCallInSrcRector::class,
        RequirePsrClockInterfaceRector::class,
        NoAssertInsideIfInFunctionalTestsRector::class,
        NoArrayAssertContainsInTestsRector::class,
        NoTypeOnlyAssertionsInTestsRector::class,
        NoExistenceOnlyAssertionsInTestsRector::class,
        NoDirectDbMutationInFunctionalTestsRector::class,
        NoDirectDispatchInFunctionalTestsRector::class,
        NoNullCoalesceNewFallbackRector::class,
    ])
    ->withConfiguredRule(YamlNoCommentsRector::class, [
        YamlNoCommentsRector::PATHS => [
            __DIR__ . '/config',
            __DIR__ . '/translations',
            __DIR__ . '/bundles',
        ],
    ]);
