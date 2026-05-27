<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Constraint;

use Egulias\EmailValidator\EmailValidator as EguliasEmailValidator;
use Egulias\EmailValidator\Validation\DNSCheckValidation;
use Symfony\Component\Validator\Constraint;
use Symfony\Component\Validator\ConstraintValidator;
use Symfony\Component\Validator\Exception\UnexpectedTypeException;
use Symfony\Component\Validator\Exception\UnexpectedValueException;

/**
 * @author Bernhard Schussek <bschussek@gmail.com>
 *
 * @SuppressWarnings(PHPMD)
 */
class EmailValidator extends ConstraintValidator
{
    private const PATTERN_HTML5 = '/^[a-zA-Z0-9.!#$%&\'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$/';

    private const PATTERN_LOOSE = '/^.+\@\S+\.\S+$/';

    private const EMAIL_PATTERNS = [
        Email::VALIDATION_MODE_LOOSE => self::PATTERN_LOOSE,
        Email::VALIDATION_MODE_HTML5 => self::PATTERN_HTML5,
    ];

    private $defaultMode;

    public function __construct(string $defaultMode = Email::VALIDATION_MODE_LOOSE)
    {
        if (!\in_array($defaultMode, Email::$validationModes, true)) {
            throw new \InvalidArgumentException('The "defaultMode" parameter value is not valid.');
        }

        $this->defaultMode = $defaultMode;
    }

    /**
     * {@inheritdoc}
     */
    public function validate($value, Constraint $constraint)
    {
        if (!$constraint instanceof Email) {
            throw new UnexpectedTypeException($constraint, Email::class);
        }

        if (null === $value || '' === $value) {
            return;
        }

        if (!is_scalar($value) && !(\is_object($value) && method_exists($value, '__toString'))) {
            throw new UnexpectedValueException($value, 'string');
        }

        $value = (string) $value;
        if ('' === $value) {
            return;
        }

        if (null !== $constraint->normalizer) {
            $value = ($constraint->normalizer)($value);
        }

        if (null === $constraint->mode) {
            $constraint->mode = $this->defaultMode;
        }

        if (!\in_array($constraint->mode, Email::$validationModes, true)) {
            throw new \InvalidArgumentException(sprintf('The "%s::$mode" parameter value is not valid.', get_debug_type($constraint)));
        }

        if (Email::VALIDATION_MODE_STRICT === $constraint->mode) {
            $strictValidator = new EguliasEmailValidator();

            if (!$strictValidator->isValid($value, new DNSCheckValidation())) {
                $this->context->buildViolation($constraint->message)
                    ->setParameter('{{ value }}', $this->formatValue($value))
                    ->setCode(Email::INVALID_FORMAT_ERROR)
                    ->addViolation();

                return;
            }
        }

        if (!preg_match(self::EMAIL_PATTERNS[Email::VALIDATION_MODE_LOOSE], $value)) {
            $this->context->buildViolation($constraint->message)
                ->setParameter('{{ value }}', $this->formatValue($value))
                ->setCode(Email::INVALID_FORMAT_ERROR)
                ->addViolation();
        }
    }
}
