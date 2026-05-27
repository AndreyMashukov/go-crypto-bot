<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Form;

use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class CryptoTradeConfigUpdateType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options): void
    {
        $builder
            ->add('usdtLimit')
            ->add('enabled')
            ->add('signalTrading')
            ->add('minPriceMinutesPeriod')
            ->add('framePeriod')
            ->add('frameInterval')
            ->add('buyPriceHistoryCheckInterval')
            ->add('buyPriceHistoryCheckPeriod')
            ->add('signalConfig', SignalConfigType::class)
        ;
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'allow_extra_fields' => true,
            'csrf_protection'    => false,
            'data_class'         => CryptoTradeConfig::class,
            'validation_groups'  => ['cryptotrade_config'],
        ]);
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }
}
