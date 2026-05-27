<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Form;

use Bundles\CryptoBotContext\Model\ProfitOption;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\NumberType;
use Symfony\Component\Form\Extension\Core\Type\TextType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class ProfitOptionType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options)
    {
        $builder
            ->add('index', NumberType::class)
            ->add('optionUnit', TextType::class)
            ->add('optionValue', NumberType::class)
            ->add('optionPercent', NumberType::class)
            ->add('isTriggerOption')
        ;
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'csrf_protection' => false,
            'data_class'      => ProfitOption::class,
        ]);
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }
}
