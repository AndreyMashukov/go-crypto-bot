<?php

declare(strict_types=1);

namespace Bundles\UserContext\Form;

use App\Form\ExtendedEntityType;
use Bundles\OxaPayContext\Entity\PromoCode;
use Bundles\UserContext\Model\Register;
use Doctrine\ORM\EntityManagerInterface;
use Doctrine\ORM\Mapping\ClassMetadata;
use Symfony\Bridge\Doctrine\Form\ChoiceList\IdReader;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class RegisterType extends AbstractType
{
    public function __construct(private readonly EntityManagerInterface $entityManager)
    {
    }

    public function buildForm(FormBuilderInterface $builder, array $options)
    {
        $builder
            ->add('nickname')
            ->add('email')
            ->add('secret')
            ->add('promoCode', ExtendedEntityType::class, [
                'class'     => PromoCode::class,
                'id_reader' => $this->getPromoCodeIdReader(),
            ])
        ;
    }

    public function configureOptions(OptionsResolver $resolver)
    {
        $resolver->setDefaults([
            'csrf_protection' => false,
            'data_class'      => Register::class,
        ]);
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }

    private function getPromoCodeIdReader(): IdReader
    {
        $fakeMetadata = new ClassMetadata(PromoCode::class);
        $fakeMetadata->setIdentifier(['code']);
        $fakeMetadata->fieldMappings['code']['type'] = 'string';
        $reflectionProperty                          = new \ReflectionProperty(PromoCode::class, 'code');
        $fakeMetadata->reflFields['code'] = $reflectionProperty;

        return new IdReader($this->entityManager, $fakeMetadata);
    }
}
