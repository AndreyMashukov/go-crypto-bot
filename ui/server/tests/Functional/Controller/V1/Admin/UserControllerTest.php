<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Tests\Functional\Controller\V1\Admin;

use App\Tests\RestTestCase;
use Symfony\Component\HttpFoundation\Request;

/**
 * @group functional
 */
class UserControllerTest extends RestTestCase
{
    /**
     * Should allow to do CRUD operations.
     */
    public function testShouldAllowToDoCrudOperations(): void
    {
        $list = $this->deserialize($this->apiRequest($this->getUrl('v1_admin_list', [
            'page'  => 1,
            'limit' => 250,
        ])));
        $this->assertJsonSnapshot($list, 'list');

        $updated = $this->deserialize($this->apiRequest($this->getUrl('v1_admin_patch', [
            'user'     => $list['data'][0]['id'],
        ]), Request::METHOD_PATCH, [
            'email'    => 'test_email',
            'phone'    => 'test_phone',
        ]));
        $this->assertJsonSnapshot($updated, 'updated');

        $freeze = $this->deserialize($this->apiRequest($this->getUrl('v1_admin_put_freeze_switch', [
            'user' => $list['data'][0]['id'],
        ]), Request::METHOD_PUT));
        $this->assertJsonSnapshot($freeze, 'freeze');
    }
}
