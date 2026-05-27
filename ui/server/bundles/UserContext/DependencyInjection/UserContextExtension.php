<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\DependencyInjection;

use Symfony\Component\Config\FileLocator;
use Symfony\Component\DependencyInjection\ContainerBuilder;
use Symfony\Component\DependencyInjection\Extension\Extension;
use Symfony\Component\DependencyInjection\Extension\PrependExtensionInterface;
use Symfony\Component\DependencyInjection\Loader\YamlFileLoader;
use Symfony\Component\Yaml\Yaml;

final class UserContextExtension extends Extension implements PrependExtensionInterface
{
    /**
     * {@inheritDoc}
     *
     * @param array<array> $configs
     *
     * @throws \Exception
     */
    public function load(array $configs, ContainerBuilder $container): void
    {
        $configuration = new Configuration();
        $this->processConfiguration($configuration, $configs);

        $loader = new YamlFileLoader(
            $container,
            new FileLocator(__DIR__ . '/../Resources/config/')
        );
        $loader->load('services.yaml');

        if ('test' === $container->getParameter('kernel.environment')) {
            $loader->load('services_test.yaml');
        }
    }

    public function prepend(ContainerBuilder $container)
    {
        $twigConfig = Yaml::parseFile(__DIR__ . '/../Resources/config/packages/twig.yaml');
        $container->prependExtensionConfig('twig', $twigConfig['twig']);

        $doctrineConfig = Yaml::parseFile(__DIR__ . '/../Resources/config/packages/doctrine.yaml');
        $container->prependExtensionConfig('doctrine', $doctrineConfig['doctrine']);
    }
}
