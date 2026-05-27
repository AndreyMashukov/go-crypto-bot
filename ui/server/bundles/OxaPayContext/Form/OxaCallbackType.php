<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\OxaPayContext\Form;

use Bundles\OxaPayContext\Model\OxaCallback;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\CheckboxType;
use Symfony\Component\Form\Extension\Core\Type\NumberType;
use Symfony\Component\Form\Extension\Core\Type\TextType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class OxaCallbackType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options): void
    {
        $builder
            ->add('status', TextType::class)
            ->add('trackId', NumberType::class)
            ->add('amount', NumberType::class)
            ->add('currency', TextType::class)
            ->add('feePaidByPayer', CheckboxType::class)
            ->add('underPaidCover', NumberType::class)
            ->add('email', TextType::class)
            ->add('orderId', TextType::class)
            ->add('description', TextType::class)
            ->add('date', NumberType::class)
            ->add('payDate', NumberType::class)
            ->add('type', TextType::class)
            ->add('payAmount', NumberType::class)
            ->add('payCurrency', TextType::class)
            ->add('price', NumberType::class)
        ;
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'csrf_protection'    => false,
            'data_class'         => OxaCallback::class,
            'allow_extra_fields' => true,
        ]);
    }

    public function getBlockPrefix()
    {
        return '';
    }
}
