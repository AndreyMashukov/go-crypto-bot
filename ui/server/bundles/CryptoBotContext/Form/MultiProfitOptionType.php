<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Form;

use Bundles\CryptoBotContext\Model\MultiProfitOption;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\CollectionType;
use Symfony\Component\Form\Extension\Core\Type\NumberType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class MultiProfitOptionType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options)
    {
        $builder
            ->add('orderId', NumberType::class)
            ->add('profitOptions', CollectionType::class, [
                'allow_add'    => true,
                'allow_delete' => true,
                'entry_type'   => ProfitOptionType::class,
            ]);
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'csrf_protection' => false,
            'data_class'      => MultiProfitOption::class,
        ]);
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }
}
