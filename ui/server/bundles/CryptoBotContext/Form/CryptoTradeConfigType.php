<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Form;

use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\CollectionType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class CryptoTradeConfigType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options): void
    {
        $builder
            ->add('symbol')
            ->add('usdtLimit')
            ->add('enabled')
            ->add('signalTrading')
            ->add('minPriceMinutesPeriod')
            ->add('framePeriod')
            ->add('frameInterval')
            ->add('buyPriceHistoryCheckInterval')
            ->add('buyPriceHistoryCheckPeriod')
            ->add('extraChargeOptions', CollectionType::class, [
                'allow_add'     => true,
                'allow_delete'  => true,
                'entry_type'    => ExtraChargeOptionType::class,
                'entry_options' => [
                    'data_class' => null,
                ],
            ])
            ->add('profitOptions', CollectionType::class, [
                'allow_add'     => true,
                'allow_delete'  => true,
                'entry_type'    => ProfitOptionType::class,
                'entry_options' => [
                    'data_class' => null,
                ],
            ])
        ;
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'allow_extra_fields' => true,
            'csrf_protection'    => false,
            'data_class'         => CryptoTradeConfig::class,
        ]);
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }
}
