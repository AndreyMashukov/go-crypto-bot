<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Functional\Controller\PublicRoute;

use App\Tests\RestTestCase;

/**
 * @group functional
 */
class CommissionControllerTest extends RestTestCase
{
    /**
     * Should allow to get commission info.
     */
    public function testShouldAllowToGetCommissionInfo(): void
    {
        $json              = $this->deserialize($this->apiPublicRequest($this->getUrl('public_commission_info')));
        $json[3]['volume'] = '{max_float}';
        $this->assertJsonSnapshot($json);
    }
}
