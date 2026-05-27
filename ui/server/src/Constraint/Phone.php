<?php

declare(strict_types=1);

namespace App\Constraint;

use Symfony\Component\Validator\Constraint;

class Phone extends Constraint
{
    public $message = 'Номер {{ value }} невалидный телефонный номер.';

    #[\Override]
    public function validatedBy(): string
    {
        return static::class . 'Validator';
    }
}
