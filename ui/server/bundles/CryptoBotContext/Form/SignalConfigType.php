<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Form;

use Bundles\CryptoBotContext\Entity\Embedded\SignalConfig;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\ChoiceType;
use Symfony\Component\Form\Extension\Core\Type\NumberType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class SignalConfigType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options)
    {
        $builder
            ->add('ratingFilter')
            ->add('avgBuyFilter')
            ->add('avgSellFilter')
            ->add('avgBuyCorrection')
            ->add('avgSellCorrection')
            ->add('percentFilter', NumberType::class, [
                'empty_data' => '0.50',
            ])
            ->add('signalPeriodDays', NumberType::class, [
                'empty_data' => '1',
            ])
            ->add('sellPriceCorrectionMode', ChoiceType::class, [
                'choices' => [
                    SignalConfig::SELL_PRICE_CORRECTION_MODE_EQUAL => SignalConfig::SELL_PRICE_CORRECTION_MODE_EQUAL,
                    SignalConfig::SELL_PRICE_CORRECTION_MODE_MAX   => SignalConfig::SELL_PRICE_CORRECTION_MODE_MAX,
                    SignalConfig::SELL_PRICE_CORRECTION_MODE_MIN   => SignalConfig::SELL_PRICE_CORRECTION_MODE_MIN,
                ],
            ])
        ;
    }

    public function configureOptions(OptionsResolver $resolver)
    {
        $resolver->setDefaults([
            'data_class'      => SignalConfig::class,
            'csrf_protection' => false,
        ]);
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }
}
