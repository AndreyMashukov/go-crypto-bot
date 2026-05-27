<?php
namespace App\Controller\PublicRoute;

use Bundles\CryptoBotContext\Form\SignalType;
use Bundles\CryptoBotContext\Model\Signal;
use Bundles\CryptoBotContext\Service\SignalHandler;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Exception\AccessDeniedHttpException;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;

class InternalController extends AbstractFOSRestController
{
    public function __construct(private readonly string $internalToken, private readonly SignalHandler $signalHandler)
    {
    }

    public function postSignal(Request $request)
    {
        $this->verifyRequest($request);

        $form = $this->createForm(SignalType::class, new Signal(), [
            'method' => Request::METHOD_POST,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            $errors = $form->getErrors(true);
            throw new BadRequestHttpException((string) $errors[0]->getCause());
        }

        $signal = $form->getData();
        $this->signalHandler->handleSignal($signal);
    }

    private function verifyRequest(Request $request): void
    {
        if (!$request->headers->has('Crypto-Internal-Token')) {
            throw new AccessDeniedHttpException('Access Denied');
        }

        $token = $request->headers->get('Crypto-Internal-Token');
        if ($token !== $this->internalToken) {
            throw new AccessDeniedHttpException('Access Denied');
        }
    }
}
