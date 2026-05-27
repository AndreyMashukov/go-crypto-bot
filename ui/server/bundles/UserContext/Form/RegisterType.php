<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

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
    private EntityManagerInterface $entityManager;

    public function __construct(EntityManagerInterface $entityManager)
    {
        $this->entityManager = $entityManager;
    }

    /**
     * {@inheritdoc}
     */
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

    /**
     * {@inheritdoc}
     */
    public function configureOptions(OptionsResolver $resolver)
    {
        $resolver->setDefaults([
            'csrf_protection' => false,
            'data_class'      => Register::class,
        ]);
    }

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
        $reflectionProperty->setAccessible(true);
        $fakeMetadata->reflFields['code'] = $reflectionProperty;

        return new IdReader($this->entityManager, $fakeMetadata);
    }
}
