<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Form;

use Bundles\CryptoBotContext\Model\TradeActionCondition;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\TextType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class TradeActionConditionChildrenType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options)
    {
        $builder
            ->add('symbol', TextType::class)
            ->add('parameter', TextType::class)
            ->add('condition', TextType::class)
            ->add('value', TextType::class)
            ->add('type', TextType::class)
        ;
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'csrf_protection' => false,
            'data_class'      => TradeActionCondition::class,
        ]);
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }
}
