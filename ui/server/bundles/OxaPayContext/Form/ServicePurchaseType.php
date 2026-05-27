<?php

declare(strict_types=1);

namespace Bundles\OxaPayContext\Form;

use Bundles\OxaPayContext\Model\ServicePurchase;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\TextType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class ServicePurchaseType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options)
    {
        $builder
            ->add('code', TextType::class)
        ;
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'data_class'      => ServicePurchase::class,
            'csrf_protection' => false,
        ]);
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }
}
