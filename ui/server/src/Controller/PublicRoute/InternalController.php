<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\PublicRoute;

use Bundles\CryptoBotContext\Form\SignalType;
use Bundles\CryptoBotContext\Model\Signal;
use Bundles\CryptoBotContext\Service\SignalHandler;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Symfony\Component\Form\FormError;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Exception\AccessDeniedHttpException;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;

/**
 * @Rest\Route("/public/internal", name="public_internal_")
 */
class InternalController extends AbstractFOSRestController
{
    private string $internalToken;

    private SignalHandler $signalHandler;

    public function __construct(string $internalToken, SignalHandler $signalHandler)
    {
        $this->internalToken = $internalToken;
        $this->signalHandler = $signalHandler;
    }

    /**
     * @Rest\Route("/signal", methods={"POST"}, name="signal")
     * @Rest\View
     *
     * @param Request $request
     */
    public function postSignalAction(Request $request)
    {
        $this->verifyRequest($request);

        $form = $this->createForm(SignalType::class, new Signal(), [
            'method' => Request::METHOD_POST,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            /** @var FormError[] $errors */
            $errors = $form->getErrors(true);
            throw new BadRequestHttpException((string) $errors[0]->getCause());
        }

        /** @var Signal $callback */
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
