<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\PublicRoute;

use App\Service\IpExtractor;
use Bundles\CryptoBotContext\Form\ErrorCallbackType;
use Bundles\CryptoBotContext\Model\ErrorCallback;
use Bundles\CryptoBotContext\Repository\CryptoBotRepository;
use Bundles\CryptoBotContext\Service\DeployDomain;
use Bundles\OxaPayContext\Form\OxaCallbackType;
use Bundles\OxaPayContext\Model\OxaCallback;
use Bundles\OxaPayContext\Service\PaymentManager;
use Bundles\TgBotContext\Form\OrderMessageType;
use Bundles\TgBotContext\MessageBuilder;
use Bundles\TgBotContext\Model\ChatConfiguration;
use Bundles\TgBotContext\Model\OrderMessage;
use Bundles\TgBotContext\Service\AlertService;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;
use Symfony\Component\HttpKernel\Exception\ServiceUnavailableHttpException;
use TgBotApi\BotApiBase\BotApi;
use TgBotApi\BotApiBase\BotApiComplete;

/**
 * @Rest\Route("/public/callback", name="public_callback_")
 */
class CallbackController extends AbstractFOSRestController
{
    private PaymentManager $paymentManager;

    private MessageBuilder $messageBuilder;

    private BotApi $botApi;

    private CryptoBotRepository $cryptoBotRepository;

    private IpExtractor $ipExtractor;

    private DeployDomain $deployDomain;

    private AlertService $alertService;

    public function __construct(
        PaymentManager $paymentManager,
        MessageBuilder $messageBuilder,
        BotApiComplete $botApi,
        CryptoBotRepository $cryptoBotRepository,
        IpExtractor $ipExtractor,
        DeployDomain $deployDomain,
        AlertService $alertService
    ) {
        $this->paymentManager      = $paymentManager;
        $this->messageBuilder      = $messageBuilder;
        $this->botApi              = $botApi;
        $this->cryptoBotRepository = $cryptoBotRepository;
        $this->ipExtractor         = $ipExtractor;
        $this->deployDomain        = $deployDomain;
        $this->alertService        = $alertService;
    }

    /**
     * @Rest\Route("/oxa", methods={"POST"}, name="oxa")
     * @Rest\View
     *
     * @param Request $request
     *
     * @return Response
     */
    public function postOxaAction(Request $request): Response
    {
        $form = $this->createForm(OxaCallbackType::class, new OxaCallback(), [
            'method' => Request::METHOD_POST,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            $this->alertService->alert('CallbackController: OXA callback error, invalid request.');
            throw new BadRequestHttpException('Invalid request.');
        }

        /** @var OxaCallback $callback */
        $callback = $form->getData();

        try {
            $this->paymentManager->validateCallback($callback);
        } catch (\BadMethodCallException $exception) {
            throw new BadRequestHttpException('Invalid request.', $exception);
        }

        return new Response('OK', 200, [
            'Content-Type' => 'text/plain',
        ]);
    }

    /**
     * @Rest\Route("/telegram", methods={"POST"}, name="telegram")
     * @Rest\View
     *
     * @param Request $request
     *
     * @return Response
     */
    public function postTelegramAction(Request $request): Response
    {
        $form = $this->createForm(OrderMessageType::class, new OrderMessage(), [
            'method' => Request::METHOD_POST,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            throw new BadRequestHttpException('Invalid request.');
        }

        /** @var OrderMessage $orderMessage */
        $orderMessage = $form->getData();

        if (!$orderMessage->getBot()->isTest()) {
            try {
                $tgMessages = $this->messageBuilder->fromOrderMessage($orderMessage);
                foreach ($tgMessages as $tgMessage) {
                    $tgMessage->send($this->botApi, new ChatConfiguration(-1002011203736));
                }
            } catch (\Exception $exception) {
                throw new ServiceUnavailableHttpException(60, 'Service unavailable', $exception);
            }
        }

        return new Response('OK', 200, [
            'Content-Type' => 'text/plain',
        ]);
    }

    /**
     * @Rest\Route("/error", methods={"POST"}, name="error")
     * @Rest\View
     *
     * @param Request $request
     *
     * @return Response
     */
    public function postErrorAction(Request $request): Response
    {
        $form = $this->createForm(ErrorCallbackType::class, new ErrorCallback(), [
            'method' => Request::METHOD_POST,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            throw new BadRequestHttpException('Invalid request.');
        }

        /** @var ErrorCallback $data */
        $data = $form->getData();

        if ($data->stop) {
            if ($data->bot->getServer()) {
                $this->deployDomain->doStop(
                    $data->bot,
                    $data->bot->getServer(),
                    '',
                    false
                );
            }
        }

        $this->alertService->alert("CallbackController: Bot #{$data->bot->getId()} User {$data->bot->getUser()->getEmail()} - {$data->errorMessage}");

        $data->bot->setErrorMessage("[{$this->ipExtractor->extractIp($request)}] {$data->errorMessage}");
        $this->cryptoBotRepository->add($data->bot, true);

        return new Response('OK', 200, [
            'Content-Type' => 'text/plain',
        ]);
    }
}
