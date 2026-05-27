<?php

declare(strict_types=1);

namespace Bundles\UserContext\Model;

interface SecureDataInterface
{
    public function getSecret(): ?string;

    public function getSecureString(): string;
}
