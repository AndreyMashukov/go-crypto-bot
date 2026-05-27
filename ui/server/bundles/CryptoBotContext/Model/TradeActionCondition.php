<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Model;

use Symfony\Component\Validator\Constraints as Assert;
use Symfony\Component\Validator\Context\ExecutionContextInterface;

class TradeActionCondition
{
    public const ALLOWED_PARAMETERS = [
        'price',
        'daily_percent',
        'position_time_minutes',
        'extra_orders_today',
        'has_signal',
        'sentiment_label',
        'sentiment_score',
    ];

    /**
     * @var null|string
     */
    public ?string $symbol = null;

    /**
     * @var null|string
     */
    public ?string $parameter = null;

    /**
     * @var null|string
     */
    public ?string $condition = null;

    /**
     * @var null|string
     */
    public ?string $value = null;

    /**
     * @var null|string
     */
    public ?string $type = null;

    /**
     * @Assert\NotNull
     * @Assert\Valid
     *
     * @var array
     */
    public array $children = [];

    /**
     * @Assert\Callback
     *
     * @param mixed $payload
     */
    public function validateSymbol(ExecutionContextInterface $context, $payload): bool
    {
        if (\count($this->children) > 0) {
            return true;
        }

        if (null === $this->symbol) {
            $context->buildViolation('This value should not be null.')
                ->atPath('symbol')
                ->setCode(Assert\NotNull::IS_NULL_ERROR)
                ->addViolation();

            return false;
        }

        return true;
    }

    /**
     * @Assert\Callback
     *
     * @param mixed $payload
     */
    public function validateParameter(ExecutionContextInterface $context, $payload): bool
    {
        if (\count($this->children) > 0) {
            return true;
        }

        if (null === $this->parameter) {
            $context->buildViolation('This value should not be null.')
                ->atPath('parameter')
                ->setCode(Assert\NotNull::IS_NULL_ERROR)
                ->addViolation();

            return false;
        }

        if (!\in_array($this->parameter, self::ALLOWED_PARAMETERS, true)) {
            $context->buildViolation('The value you selected is not a valid choice.')
                ->atPath('parameter')
                ->setCode(Assert\Choice::NO_SUCH_CHOICE_ERROR)
                ->addViolation();

            return false;
        }

        return true;
    }

    /**
     * @Assert\Callback
     *
     * @param mixed $payload
     */
    public function validateCondition(ExecutionContextInterface $context, $payload): bool
    {
        if (\count($this->children) > 0) {
            return true;
        }

        if (null === $this->condition) {
            $context->buildViolation('This value should not be null.')
                ->atPath('condition')
                ->setCode(Assert\NotNull::IS_NULL_ERROR)
                ->addViolation();

            return false;
        }

        $availableConditions = ['lt', 'lte', 'gt', 'gte', 'eq', 'neq'];
        if ('has_signal' === $this->parameter) {
            $availableConditions = ['eq', 'neq'];
        }

        if (!\in_array($this->condition, $availableConditions, true)) {
            $context->buildViolation('The value you selected is not a valid choice.')
                ->atPath('condition')
                ->setCode(Assert\Choice::NO_SUCH_CHOICE_ERROR)
                ->addViolation();

            return false;
        }

        return true;
    }

    /**
     * @Assert\Callback
     *
     * @param mixed $payload
     */
    public function validateValue(ExecutionContextInterface $context, $payload): bool
    {
        if (\count($this->children) > 0) {
            return true;
        }

        if (null === $this->value) {
            $context->buildViolation('This value should not be null.')
                ->atPath('value')
                ->setCode(Assert\NotNull::IS_NULL_ERROR)
                ->addViolation();

            return false;
        }

        return true;
    }
}
