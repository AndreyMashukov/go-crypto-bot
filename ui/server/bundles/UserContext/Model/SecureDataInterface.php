<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Model;

interface SecureDataInterface
{
    public function getSecret(): ?string;

    public function getSecureString(): string;
}
