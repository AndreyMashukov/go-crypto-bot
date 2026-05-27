<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\V1\Admin;

use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Form\UserUpdateType;
use Bundles\UserContext\Repository\UserRepository;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Knp\Component\Pager\Pagination\PaginationInterface;
use Knp\Component\Pager\PaginatorInterface;
use Sensio\Bundle\FrameworkExtraBundle\Configuration\IsGranted;
use Symfony\Component\HttpFoundation\Request;

/**
 * @Rest\Route("/v1/admin", name="v1_admin_")
 */
class UserController extends AbstractFOSRestController
{
    private UserRepository $repository;

    private PaginatorInterface $paginator;

    public function __construct(
        UserRepository $repository,
        PaginatorInterface $paginator
    ) {
        $this->repository = $repository;
        $this->paginator  = $paginator;
    }

    /**
     * @Rest\Route("/list", name="list", methods={"GET"})
     * @Rest\View(serializerGroups={"admin", "knp_basic"})
     * @IsGranted("ROLE_ADMIN")
     *
     * @param Request $request
     *
     * @return PaginationInterface
     */
    public function getListAction(Request $request): PaginationInterface
    {
        $page  = $request->get('page', 1);
        $limit = min((int) $request->get('limit', 50), 200);

        return $this->paginator->paginate($this->repository->getListQB(), $page, $limit);
    }

    /**
     * @Rest\Route("/{user}", name="patch", methods={"PATCH"})
     * @Rest\View(serializerGroups={"admin"})
     *
     * @param Request $request
     * @param User    $user
     *
     * @return array|User
     * @IsGranted("ROLE_ADMIN")
     */
    public function patchAction(Request $request, User $user)
    {
        $form = $this->createForm(UserUpdateType::class, $user, [
            'method' => Request::METHOD_PATCH,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            return [
                'form' => $form,
            ];
        }

        $data = $form->getData();
        $this->repository->add($data, true);

        return $data;
    }

    /**
     * @Rest\Route("/{user}/freeze/switch", name="put_freeze_switch", methods={"PUT"})
     * @Rest\View(serializerGroups={"admin"})
     *
     * @param User $user
     * @IsGranted("ROLE_ADMIN")
     *
     * @return User
     */
    public function putFreezeSwitchAction(User $user): User
    {
        $user->setFreeze(!$user->isFreeze());
        $this->repository->add($user, true);

        return $user;
    }
}
