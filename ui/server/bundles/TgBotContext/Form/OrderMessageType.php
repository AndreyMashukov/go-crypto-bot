<?php

declare(strict_types=1);

namespace Bundles\TgBotContext\Form;

use App\Form\ExtendedEntityType;
use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\TgBotContext\Model\OrderMessage;
use Doctrine\ORM\EntityManagerInterface;
use Doctrine\ORM\Mapping\ClassMetadata;
use Symfony\Bridge\Doctrine\Form\ChoiceList\IdReader;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class OrderMessageType extends AbstractType
{
    public function __construct(private readonly EntityManagerInterface $entityManager)
    {
    }

    public function buildForm(FormBuilderInterface $builder, array $options): void
    {
        $builder
            ->add('bot', ExtendedEntityType::class, [
                'class'     => CryptoBot::class,
                'id_reader' => $this->getBotIdReader(),
            ])
            ->add('dateTime')
            ->add('symbol')
            ->add('amount')
            ->add('price')
            ->add('operation')
            ->add('details')
        ;
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'data_class'      => OrderMessage::class,
            'csrf_protection' => false,
        ]);
    }

    #[\Override]
    public function getBlockPrefix()
    {
        return '';
    }

    private function getBotIdReader(): IdReader
    {
        $fakeMetadata = new ClassMetadata(CryptoBot::class);
        $fakeMetadata->setIdentifier(['uuid']);
        $fakeMetadata->fieldMappings['uuid']['type'] = 'string';
        $reflectionProperty                          = new \ReflectionProperty(CryptoBot::class, 'uuid');
        $fakeMetadata->reflFields['uuid'] = $reflectionProperty;

        return new IdReader($this->entityManager, $fakeMetadata);
    }
}
