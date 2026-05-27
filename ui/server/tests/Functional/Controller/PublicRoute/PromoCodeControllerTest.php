<?php
namespace Functional\Controller\PublicRoute;

use App\Tests\RestTestCase;

class PromoCodeControllerTest extends RestTestCase
{
    public function testShouldAllowToValidatePromoCode(): void
    {
        $json = $this->deserialize($this->apiPublicRequest($this->getUrl('public_promocode_test', [
            'code' => 'TEST123',
        ])));
        $this->assertJsonSnapshot($json);
    }
}
