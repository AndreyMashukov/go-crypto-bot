<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\PublicRoute;

use Bundles\UserContext\Form\RegisterType;
use Bundles\UserContext\Model\Register;
use Bundles\UserContext\Service\EmailSender;
use Bundles\UserContext\Service\RegistrationService;
use Bundles\UserContext\Traits\SecureValidation;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Psr\Log\LoggerInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Exception\AccessDeniedHttpException;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;

/**
 * @Rest\Route("/public/register", name="public_registration_")
 */
class RegistrationController extends AbstractFOSRestController
{
    use SecureValidation;

    private RegistrationService $registrationService;

    private string $environment;

    private LoggerInterface $logger;

    private EmailSender $emailSender;

    public function __construct(string $environment, RegistrationService $registrationService, EmailSender $emailSender, LoggerInterface $logger)
    {
        $this->registrationService = $registrationService;
        $this->environment         = $environment;
        $this->logger              = $logger;
        $this->emailSender         = $emailSender;
    }

    /**
     * @Rest\Route("/code", methods={"POST"}, name="code")
     * @Rest\View
     *
     * @param Request $request
     *
     * @throws \Exception
     *
     * @return array|void
     */
    public function postAction(Request $request)
    {
        $form = $this->createForm(RegisterType::class, new Register($request->getLocale()), [
            'method' => Request::METHOD_POST,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            return [
                'form' => $form,
            ];
        }

        /** @var Register $register */
        $register = $form->getData();

        if (!$this->validateSecret($register)) {
            throw new AccessDeniedHttpException('time_based_error');
        }

        try {
            [$code, $user] = $this->registrationService->registerByEmail($register);
        } catch (\BadMethodCallException $exception) {
            throw new BadRequestHttpException($exception->getMessage(), $exception);
        }

        if ('dev' === $this->environment) {
            $this->logger->debug("code is {$code}");

            return;
        }

        $this->emailSender->send(14, [
            'code' => $code,
        ], $user);
    }
}
