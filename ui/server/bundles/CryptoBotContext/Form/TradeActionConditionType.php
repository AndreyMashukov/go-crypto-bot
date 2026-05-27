<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

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

    public function getParent()
    {
        return TradeActionConditionChildrenType::class;
    }

    public function getBlockPrefix()
    {
        return '';
    }
}
