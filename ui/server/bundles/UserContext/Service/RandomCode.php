<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Service;

class RandomCode
{
    /**
     * @param int $bites
     *
     * @throws \Exception
     *
     * @return string
     */
    public static function string(int $bites): string
    {
        return bin2hex(random_bytes($bites));
    }

    /**
     * @param int $bites
     *
     * @throws \Exception
     *
     * @return string
     */
    public function generate(int $bites): string
    {
        return self::string($bites);
    }
}
