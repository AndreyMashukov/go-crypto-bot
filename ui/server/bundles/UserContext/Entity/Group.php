<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Entity;

use Doctrine\ORM\Mapping as ORM;
use Pd\UserBundle\Model\Group as BaseGroup;
use Symfony\Bridge\Doctrine\Validator\Constraints\UniqueEntity;

/**
 * @ORM\Table(name="user_group")
 * @ORM\Entity
 * @UniqueEntity(fields="name", message="group_already_taken")
 */
class Group extends BaseGroup
{
}
