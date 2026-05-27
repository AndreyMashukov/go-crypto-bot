<?php
namespace App\Constraint;

use Symfony\Component\Validator\Constraint;
use Symfony\Component\Validator\ConstraintValidator;
use Symfony\Component\Validator\Exception\UnexpectedTypeException;

class PhoneValidator extends ConstraintValidator
{
    private const string MASK = '/^7\d{10}$/';

    public function validate(mixed $value, Constraint $constraint): void
    {
        if (!$constraint instanceof Phone) {
            throw new UnexpectedTypeException($constraint, __NAMESPACE__ . '\Phone');
        }

        if (null === $value || '' === $value) {
            return;
        }

        if (!preg_match(self::MASK, (string) $value)) {
            $this->context->buildViolation($constraint->message)
                ->setParameter('{{ value }}', $this->formatValue($value))
                ->addViolation();
        }
    }
}
