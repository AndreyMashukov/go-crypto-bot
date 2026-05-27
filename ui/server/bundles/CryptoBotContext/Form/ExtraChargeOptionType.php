<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Form;

use Bundles\CryptoBotContext\Model\ExtraChargeOption;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\NumberType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class ExtraChargeOptionType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options)
    {
        $builder
            ->add('index', NumberType::class)
            ->add('percent', NumberType::class)
            ->add('amountUsdt', NumberType::class)
        ;
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'csrf_protection' => false,
            'data_class'      => ExtraChargeOption::class,
        ]);
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }
}
