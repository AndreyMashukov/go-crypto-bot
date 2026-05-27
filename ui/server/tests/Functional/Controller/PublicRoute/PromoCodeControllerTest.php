<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Functional\Controller\PublicRoute;

use App\Tests\RestTestCase;

/**
 * @group functional
 */
class PromoCodeControllerTest extends RestTestCase
{
    /**
     * Should allow to validate promo code.
     */
    public function testShouldAllowToValidatePromoCode(): void
    {
        $json = $this->deserialize($this->apiPublicRequest($this->getUrl('public_promocode_test', [
            'code' => 'TEST123',
        ])));
        $this->assertJsonSnapshot($json);
    }
}
