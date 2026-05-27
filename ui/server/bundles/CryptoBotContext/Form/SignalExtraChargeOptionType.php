<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Form;

use Bundles\CryptoBotContext\Model\SignalExtraChargeOption;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\NumberType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class SignalExtraChargeOptionType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options)
    {
        $builder
            ->add('index', NumberType::class)
            ->add('percent', NumberType::class)
            ->add('buyPrice', NumberType::class)
            ->add('budgetPercentage', NumberType::class)
        ;
    }

    public function configureOptions(OptionsResolver $resolver)
    {
        $resolver->setDefaults([
            'csrf_protection' => false,
            'data_class'      => SignalExtraChargeOption::class,
        ]);
    }

    public function getBlockPrefix()
    {
        return '';
    }
}
