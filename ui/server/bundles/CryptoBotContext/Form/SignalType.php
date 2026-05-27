<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Form;

use Bundles\CryptoBotContext\Model\Signal;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\CollectionType;
use Symfony\Component\Form\Extension\Core\Type\NumberType;
use Symfony\Component\Form\Extension\Core\Type\TextType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class SignalType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options)
    {
        $builder
            ->add('exchange', TextType::class)
            ->add('symbol', TextType::class)
            ->add('buyPrice', NumberType::class)
            ->add('percent', NumberType::class)
            ->add('periodDays', NumberType::class)
            ->add('profitOptions', CollectionType::class, [
                'allow_add'    => true,
                'allow_delete' => true,
                'entry_type'   => SignalProfitOptionType::class,
            ])
            ->add('extraChargeOptions', CollectionType::class, [
                'allow_add'    => true,
                'allow_delete' => true,
                'entry_type'   => SignalExtraChargeOptionType::class,
            ])
            ->add('expireTimestamp', NumberType::class)
        ;
    }

    public function configureOptions(OptionsResolver $resolver)
    {
        $resolver->setDefaults([
            'csrf_protection'    => false,
            'data_class'         => Signal::class,
            'allow_extra_fields' => true,
        ]);
    }

    public function getBlockPrefix()
    {
        return '';
    }
}
