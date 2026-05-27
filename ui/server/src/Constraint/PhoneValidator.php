<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Constraint;

use Symfony\Component\Validator\Constraint;
use Symfony\Component\Validator\ConstraintValidator;
use Symfony\Component\Validator\Exception\UnexpectedTypeException;

/**
 * Class PhoneValidator.
 */
class PhoneValidator extends ConstraintValidator
{
    /**
     * Pattern for mask of phone number.
     */
    private const MASK = '/^7[0-9]{10}$/';

    /**
     * @param mixed      $value
     * @param Constraint $constraint
     */
    public function validate($value, Constraint $constraint)
    {
        if (!$constraint instanceof Phone) {
            throw new UnexpectedTypeException($constraint, __NAMESPACE__ . '\Phone');
        }

        if (null === $value || '' === $value) {
            return;
        }

        if (!preg_match(self::MASK, $value)) {
            $this->context->buildViolation($constraint->message)
                ->setParameter('{{ value }}', $this->formatValue($value))
                ->addViolation();
        }
    }
}
