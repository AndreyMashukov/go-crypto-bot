<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Constraint;

use Symfony\Component\Validator\Constraint;

/**
 * Class Phone.
 *
 * @Annotation
 * @Target({"PROPERTY", "METHOD", "ANNOTATION"})
 */
class Phone extends Constraint
{
    /**
     * @var string
     */
    public $message = 'Номер {{ value }} невалидный телефонный номер.';

    /**
     * @return string
     */
    public function validatedBy()
    {
        return static::class . 'Validator';
    }
}
