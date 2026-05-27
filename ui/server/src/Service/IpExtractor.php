<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Service;

use Symfony\Component\HttpFoundation\Request;

class IpExtractor
{
    public function extractIp(Request $request): string
    {
        $ipAddress = explode(',', $request->headers->get('cf-connecting-ip', $request->getClientIp()));

        if ('172.17.0.1' === $ipAddress[0]) {
            $ipAddress = explode(',', $request->headers->get('x-forwarded-for', $request->getClientIp()));
        }

        return trim($ipAddress[0]);
    }
}
