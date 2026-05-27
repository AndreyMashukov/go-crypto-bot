<?php
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

class RegistrationController extends AbstractFOSRestController
{
    use SecureValidation;

    public function __construct(private string $environment, private RegistrationService $registrationService, private EmailSender $emailSender, private LoggerInterface $logger)
    {
    }

    /**
     * @throws \Exception
     *
     * @return mixed[]|null
     */
    public function post(Request $request)
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

            return null;
        }

        $this->emailSender->send(14, [
            'code' => $code,
        ], $user);
        return null;
    }
}
