<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Form;

use Doctrine\ORM\EntityManagerInterface;
use Doctrine\Persistence\ManagerRegistry;
use Doctrine\Persistence\ObjectManager;
use Symfony\Bridge\Doctrine\Form\ChoiceList\DoctrineChoiceLoader;
use Symfony\Bridge\Doctrine\Form\ChoiceList\EntityLoaderInterface;
use Symfony\Bridge\Doctrine\Form\ChoiceList\IdReader;
use Symfony\Bridge\Doctrine\Form\Type\EntityType;
use Symfony\Component\Form\ChoiceList\ChoiceList;
use Symfony\Component\Form\Exception\RuntimeException;
use Symfony\Component\OptionsResolver\Options;
use Symfony\Component\OptionsResolver\OptionsResolver;

class ExtendedEntityType extends EntityType
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry);
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $choiceLoader = function (Options $options) {
            // Unless the choices are given explicitly, load them on demand
            if (null === $options['choices']) {
                // If there is no QueryBuilder we can safely cache
                $vary = [$options['em'], $options['class']];

                // also if concrete Type can return important QueryBuilder parts to generate
                // hash key we go for it as well, otherwise fallback on the instance
                if ($options['query_builder']) {
                    $vary[] = $this->getQueryBuilderPartsForCachingHash($options['query_builder']) ?? $options['query_builder'];
                }

                return ChoiceList::loader($this, new DoctrineChoiceLoader(
                    $options['em'],
                    $options['class'],
                    $options['id_reader'],
                    new class($this->registry->getManager(), $options['class']) implements EntityLoaderInterface {
                        private EntityManagerInterface $entityManager;

                        private string $entityClass;

                        public function __construct(EntityManagerInterface $entityManager, string $entityClass)
                        {
                            $this->entityManager = $entityManager;
                            $this->entityClass   = $entityClass;
                        }

                        public function getEntities()
                        {
                            return [];
                        }

                        public function getEntitiesByIds(string $identifier, array $values)
                        {
                            if (!$values) {
                                return [];
                            }

                            return $this->entityManager->getRepository($this->entityClass)->findBy([
                                $identifier => $values[0],
                            ]);
                        }
                    }
                ), $vary);
            }

            return null;
        };

        $choiceName = function (Options $options) {
            // If the object has a single-column, numeric ID, use that ID as
            // field name. We can only use numeric IDs as names, as we cannot
            // guarantee that a non-numeric ID contains a valid form name
            if ($options['id_reader'] instanceof IdReader && $options['id_reader']->isIntId()) {
                return ChoiceList::fieldName($this, [__CLASS__, 'createChoiceName']);
            }

            // Otherwise, an incrementing integer is used as name automatically
            return null;
        };

        // The choices are always indexed by ID (see "choices" normalizer
        // and DoctrineChoiceLoader), unless the ID is composite. Then they
        // are indexed by an incrementing integer.
        // Use the ID/incrementing integer as choice value.
        $choiceValue = function (Options $options) {
            // If the entity has a single-column ID, use that ID as value
            if ($options['id_reader'] instanceof IdReader && $options['id_reader']->isSingleId()) {
                return ChoiceList::value($this, [$options['id_reader'], 'getIdValue'], $options['id_reader']);
            }

            // Otherwise, an incrementing integer is used as value automatically
            return null;
        };

        $emNormalizer = function (Options $options, $em) {
            if (null !== $em) {
                if ($em instanceof ObjectManager) {
                    return $em;
                }

                return $this->registry->getManager($em);
            }

            $em = $this->registry->getManagerForClass($options['class']);

            if (null === $em) {
                throw new RuntimeException(sprintf('Class "%s" seems not to be a managed Doctrine entity. Did you forget to map it?', $options['class']));
            }

            return $em;
        };

        // Invoke the query builder closure so that we can cache choice lists
        // for equal query builders
        $queryBuilderNormalizer = function (Options $options, $queryBuilder) {
            if (\is_callable($queryBuilder)) {
                $queryBuilder = $queryBuilder($options['em']->getRepository($options['class']));
            }

            return $queryBuilder;
        };

        $resolver->setDefaults([
            'em'                        => null,
            'query_builder'             => null,
            'choices'                   => null,
            'choice_loader'             => $choiceLoader,
            'choice_label'              => ChoiceList::label($this, [__CLASS__, 'createChoiceLabel']),
            'choice_name'               => $choiceName,
            'choice_value'              => $choiceValue,
            'id_reader'                 => null, // internal
            'choice_translation_domain' => false,
        ]);

        $resolver->setRequired(['class']);

        $resolver->setNormalizer('em', $emNormalizer);
        $resolver->setNormalizer('query_builder', $queryBuilderNormalizer);
        //$resolver->setNormalizer('id_reader', $idReaderNormalizer);

        $resolver->setAllowedTypes('em', ['null', 'string', ObjectManager::class]);
    }
}
