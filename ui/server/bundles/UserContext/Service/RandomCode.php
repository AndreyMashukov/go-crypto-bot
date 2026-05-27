<?php
namespace Bundles\UserContext\Service;

class RandomCode
{
    /**
     *
     * @throws \Exception
     *
     */
    public static function string(int $bites): string
    {
        return bin2hex(random_bytes($bites));
    }

    /**
     *
     * @throws \Exception
     *
     */
    public function generate(int $bites): string
    {
        return self::string($bites);
    }
}
