<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Form;

use Bundles\CryptoBotContext\Entity\ExchangeSymbol;
use Bundles\CryptoBotContext\Model\QuickConfig;
use Symfony\Bridge\Doctrine\Form\Type\EntityType;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class QuickConfigType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options)
    {
        $builder
            ->add('exchangeSymbol', EntityType::class, [
                'class' => ExchangeSymbol::class,
            ])
            ->add('restartBot')
        ;
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'data_class'      => QuickConfig::class,
            'csrf_protection' => false,
        ]);
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }
}
