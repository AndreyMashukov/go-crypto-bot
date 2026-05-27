<?php
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

class UserController extends AbstractFOSRestController
{
    public function __construct(private readonly UserRepository $repository, private readonly PaginatorInterface $paginator)
    {
    }

    public function getList(Request $request): PaginationInterface
    {
        $page  = $request->get('page', 1);
        $limit = min((int) $request->get('limit', 50), 200);

        return $this->paginator->paginate($this->repository->getListQB(), $page, $limit);
    }

    /**
     * @return array|User
     */
    public function patch(Request $request, User $user)
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

    public function putFreezeSwitch(User $user): User
    {
        $user->setFreeze(!$user->isFreeze());
        $this->repository->add($user, true);

        return $user;
    }
}
