<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Form;

use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\CollectionType;
use Symfony\Component\Form\FormBuilderInterface;

class TradeActionConditionType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options)
    {
        $builder
            ->add('children', CollectionType::class, [
                'allow_add'    => true,
                'allow_delete' => true,
                'entry_type'   => TradeActionConditionChildrenType::class,
            ])
        ;
    }

    #[\Override]
    public function getParent()
    {
        return TradeActionConditionChildrenType::class;
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }
}
