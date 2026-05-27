<?php
namespace App\Tests\Functional\Controller\PublicRoute;

use App\Tests\RestTestCase;

class CommissionControllerTest extends RestTestCase
{
    public function testShouldAllowToGetCommissionInfo(): void
    {
        $json              = $this->deserialize($this->apiPublicRequest($this->getUrl('public_commission_info')));
        $json[3]['volume'] = '{max_float}';
        $this->assertJsonSnapshot($json);
    }
}
