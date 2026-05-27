<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\V1;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Exception\ServersIsOutOfStockException;
use Bundles\CryptoBotContext\Form\CryptoBotType;
use Bundles\CryptoBotContext\Form\ManualOrderType;
use Bundles\CryptoBotContext\Form\MultiExtraChargeType;
use Bundles\CryptoBotContext\Form\MultiProfitOptionType;
use Bundles\CryptoBotContext\Form\TradeConditionCollectionType;
use Bundles\CryptoBotContext\Model\ManualOrder;
use Bundles\CryptoBotContext\Model\MultiExtraCharge;
use Bundles\CryptoBotContext\Model\MultiProfitOption;
use Bundles\CryptoBotContext\Model\TradeConditionCollection;
use Bundles\CryptoBotContext\Repository\CryptoBotRepository;
use Bundles\CryptoBotContext\Repository\ExchangeSymbolRepository;
use Bundles\CryptoBotContext\Repository\ServerRepository;
use Bundles\CryptoBotContext\Service\CryptoBotService;
use Bundles\CryptoBotContext\Service\DeployDomain;
use Bundles\OxaPayContext\Service\CommissionService;
use Bundles\TgBotContext\Service\AlertService;
use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Exception\MaxPairLimitReachedException;
use Bundles\UserContext\Service\LimitService;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Psr\Log\LoggerInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Exception\AccessDeniedHttpException;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;
use Symfony\Component\HttpKernel\Exception\ServiceUnavailableHttpException;

/**
 * @Rest\Route("/v1/cryptobot", name="v1_cryptobot_")
 */
class CryptoBotController extends AbstractFOSRestController
{
    public const AVAILABLE_PROVIDERS = [
        CryptoBot::PROVIDER_BINANCE => true,
        CryptoBot::PROVIDER_BYBIT   => true,
    ];

    private CryptoBotRepository $repository;

    private CryptoBotService $cryptoBotService;

    private ServerRepository $serverRepository;

    private DeployDomain $deployDomain;

    private LoggerInterface $logger;

    private ExchangeSymbolRepository $symbolRepository;

    private LimitService $limitService;

    private CommissionService $commissionService;

    private AlertService $alertService;

    public function __construct(
        CryptoBotRepository $repository,
        DeployDomain $deployDomain,
        CryptoBotService $cryptoBotService,
        ServerRepository $serverRepository,
        LoggerInterface $logger,
        ExchangeSymbolRepository $symbolRepository,
        LimitService $limitService,
        CommissionService $commissionService,
        AlertService $alertService
    ) {
        $this->repository        = $repository;
        $this->cryptoBotService  = $cryptoBotService;
        $this->serverRepository  = $serverRepository;
        $this->deployDomain      = $deployDomain;
        $this->logger            = $logger;
        $this->symbolRepository  = $symbolRepository;
        $this->limitService      = $limitService;
        $this->commissionService = $commissionService;
        $this->alertService      = $alertService;
    }

    /**
     * @Rest\Route("/list", name="list", methods={"GET"})
     * @Rest\View(serializerGroups={"cryptobot", "cryptotrade_config", "server"})
     *
     * @return array
     */
    public function getListAction(): array
    {
        /** @var User $user */
        $user = $this->getUser();

        return $this->repository->findBy([
            'user' => $user,
        ]);
    }

    /**
     * @Rest\Route("/list/extended", name="list_extended", methods={"GET"})
     * @Rest\View(serializerGroups={"cryptobot", "cryptotrade_config", "server", "commission"})
     *
     * @return array
     */
    public function getListExtendedAction(): array
    {
        /** @var User $user */
        $user       = $this->getUser();
        $cryptoBots = $user->getCryptoBots();

        $extendedList = [];
        foreach ($cryptoBots as $cryptobot) {
            $commission     = $this->commissionService->getCommission($cryptobot);
            $extendedList[] = [
                'cryptobot'  => $cryptobot,
                'commission' => $commission,
            ];
        }

        return $extendedList;
    }

    /**
     * @Rest\Route("/{cryptobot}/symbol/list", name="symbol_list", methods={"GET"})
     * @Rest\View(serializerGroups={"exchange_symbol"})
     *
     * @param CryptoBot $cryptobot
     *
     * @return array
     */
    public function getSymbolListAction(CryptoBot $cryptobot): array
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        return $this->symbolRepository->findBy([
            'exchange' => $cryptobot->getProvider(),
            'enabled'  => true,
        ], ['symbol' => 'ASC']);
    }

    /**
     * @Rest\Route("/available", name="available", methods={"GET"})
     * @Rest\View
     *
     * @return array
     */
    public function getAvailableBots(): array
    {
        $user = $this->getUser();
        $bots = $this->repository->findBy([
            'user' => $user,
        ]);

        $available = self::AVAILABLE_PROVIDERS;

        foreach ($bots as $bot) {
            $available[$bot->getProvider()] = false;
        }

        return $available;
    }

    /**
     * @Rest\Route("/{cryptobot}/server/list", name="bot_server_list", methods={"GET"})
     * @Rest\View(serializerGroups={"server"})
     *
     * @return array
     */
    public function getBotServerListAction(CryptoBot $cryptobot): array
    {
        if ($cryptobot->hasDedicatedServer()) {
            $list = [
                $cryptobot->getDedicated(),
            ];
        } else {
            $list = $this->serverRepository->findAll();
        }

        $stats = $this->serverRepository->getServerStats();

        return [
            'list'    => $list,
            'servers' => $stats,
        ];
    }

    /**
     * @Rest\Route("/{cryptobot}", name="get", methods={"GET"})
     * @Rest\View(serializerGroups={"cryptobot", "cryptotrade_config", "server", "cryptobot_secured"})
     *
     * @param CryptoBot $cryptobot
     *
     * @return CryptoBot
     */
    public function getAction(CryptoBot $cryptobot): CryptoBot
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        return $cryptobot;
    }

    /**
     * @Rest\Route("/{cryptobot}/deploy", name="deploy", methods={"PUT"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     */
    public function putDeployAction(CryptoBot $cryptobot): void
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        if (!$user->hasActiveBasicSubscription() && $user->getBudget() <= 0.00) {
            throw new BadRequestHttpException('You do not have active subscription or not enough account budget, please pay for service and try again.');
        }

        if (0 === $cryptobot->getCryptoTradeConfigs()->count()) {
            throw new BadRequestHttpException('You should add at least one symbol, nothing to trade');
        }

        $cryptobot->setErrorMessage(null);
        $cryptobot->setStatus(CryptoBot::STATUS_DEPLOY);
        $this->repository->add($cryptobot, true);

        try {
            $this->deployDomain->doDeploy($cryptobot);
        } catch (ServersIsOutOfStockException|\DomainException $exception) {
            $cryptobot->setStatus(CryptoBot::STATUS_ERROR);
            $this->repository->add($cryptobot, true);
            $this->logger->error($exception->getMessage(), [
                'file' => $exception->getFile(),
                'lint' => $exception->getLine(),
            ]);
            throw new ServiceUnavailableHttpException(60, $exception->getMessage(), $exception);
        }
    }

    /**
     * @Rest\Route("/{cryptobot}/stop", name="stop", methods={"PUT"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     */
    public function putStopAction(CryptoBot $cryptobot): void
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        if (!$cryptobot->getServer()) {
            throw new BadRequestHttpException('Server is not set, can not stop the bot.');
        }

        try {
            $this->deployDomain->doStop($cryptobot, $cryptobot->getServer(), '', true);
        } catch (\BadMethodCallException|\RuntimeException $exception) {
            $this->logger->error($exception->getMessage(), [
                'file' => $exception->getFile(),
                'lint' => $exception->getLine(),
            ]);
            throw new ServiceUnavailableHttpException(60, 'Service unavailable, please try later...', $exception);
        }
    }

    /**
     * @Rest\Route("/{cryptobot}/config/sync", name="config_sync", methods={"PUT"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     */
    public function putSyncConfigAction(CryptoBot $cryptobot): void
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $this->cryptoBotService->syncConfig($cryptobot);
        $this->repository->add($cryptobot, true);
    }

    /**
     * @Rest\Route("", name="post", methods={"POST"})
     * @Rest\View(serializerGroups={"cryptobot", "cryptotrade_config", "cryptobot_secured"})
     *
     * @param Request $request
     *
     * @return array|CryptoBot
     */
    public function postAction(Request $request)
    {
        /** @var User $user */
        $user = $this->getUser();
        $form = $this->createForm(CryptoBotType::class, new CryptoBot($user), [
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

        /** @var CryptoBot $data */
        $data     = $form->getData();
        $existing = $this->repository->findOneBy([
            'provider' => $data->getProvider(),
            'user'     => $user,
        ]);

        if ($existing instanceof CryptoBot) {
            throw new BadRequestHttpException("You have already created bot instance #{$existing->getId()} for {$existing->getProvider()} API");
        }

        $this->repository->add($data, true);
        $this->alertService->alert("CryptoBotController: Bot #{$data->getId()} ({$data->getProvider()}) by User {$user->getEmail()} is created.");

        return $data;
    }

    /**
     * @Rest\Route("/{cryptobot}/multi/charge", name="put_multi_charge", methods={"PUT"})
     * @Rest\View
     *
     * @param Request   $request
     * @param CryptoBot $cryptobot
     *
     * @return array|void
     */
    public function putMultiChargeAction(Request $request, CryptoBot $cryptobot)
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $form = $this->createForm(MultiExtraChargeType::class, new MultiExtraCharge($cryptobot), [
            'method' => Request::METHOD_PUT,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            return [
                'form' => $form,
            ];
        }

        /** @var MultiExtraCharge $data */
        $data = $form->getData();

        try {
            $this->cryptoBotService->setMultiExtraCharge($data);
        } catch (\Throwable $exception) {
            $this->logger->error($exception->getMessage(), [
                'file' => $exception->getFile(),
                'lint' => $exception->getLine(),
            ]);
            throw new ServiceUnavailableHttpException(60, 'Service unavailable, please try later...', $exception);
        }
    }

    /**
     * @Rest\Route("/{cryptobot}/multi/profit", name="put_multi_profit", methods={"PUT"})
     * @Rest\View
     *
     * @param Request   $request
     * @param CryptoBot $cryptobot
     *
     * @return array|void
     */
    public function putMultiProfitAction(Request $request, CryptoBot $cryptobot)
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $form = $this->createForm(MultiProfitOptionType::class, new MultiProfitOption($cryptobot), [
            'method' => Request::METHOD_PUT,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            return [
                'form' => $form,
            ];
        }

        /** @var MultiProfitOption $data */
        $data = $form->getData();

        try {
            $this->cryptoBotService->setMultiProfitOption($data);
        } catch (\Throwable $exception) {
            $this->logger->error($exception->getMessage(), [
                'file' => $exception->getFile(),
                'lint' => $exception->getLine(),
            ]);
            throw new ServiceUnavailableHttpException(60, 'Service unavailable, please try later...', $exception);
        }
    }

    /**
     * @Rest\Route("/{cryptobot}/buy/conditions", name="put_buy_conditions", methods={"PUT"})
     * @Rest\View
     *
     * @param Request   $request
     * @param CryptoBot $cryptobot
     *
     * @return array|void
     */
    public function putBuyConditionsAction(Request $request, CryptoBot $cryptobot)
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $form = $this->createForm(TradeConditionCollectionType::class, new TradeConditionCollection($cryptobot), [
            'method' => Request::METHOD_PUT,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            return [
                'form' => $form,
            ];
        }

        /** @var TradeConditionCollection $data */
        $data = $form->getData();

        try {
            $config = $cryptobot->getSymbolConfig($data->symbol);
            if ($config instanceof CryptoTradeConfig) {
                $config->setBuyConditions($data->conditions->toArray());
                $this->cryptoBotService->updateOneTradeLimit($data->cryptobot, $config);
                $this->repository->add($cryptobot, true);
            }
        } catch (\Throwable $exception) {
            $this->logger->error($exception->getMessage(), [
                'file' => $exception->getFile(),
                'lint' => $exception->getLine(),
            ]);
            throw new ServiceUnavailableHttpException(60, 'Service unavailable, please try later...', $exception);
        }
    }

    /**
     * @Rest\Route("/{cryptobot}/sell/conditions", name="put_sell_conditions", methods={"PUT"})
     * @Rest\View
     *
     * @param Request   $request
     * @param CryptoBot $cryptobot
     *
     * @return array|void
     */
    public function putSellConditionsAction(Request $request, CryptoBot $cryptobot)
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $form = $this->createForm(TradeConditionCollectionType::class, new TradeConditionCollection($cryptobot), [
            'method' => Request::METHOD_PUT,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            return [
                'form' => $form,
            ];
        }

        /** @var TradeConditionCollection $data */
        $data = $form->getData();

        try {
            $config = $cryptobot->getSymbolConfig($data->symbol);
            if ($config instanceof CryptoTradeConfig) {
                $config->setSellConditions($data->conditions->toArray());
                $this->cryptoBotService->updateOneTradeLimit($data->cryptobot, $config);
                $this->repository->add($cryptobot, true);
            }
        } catch (\Throwable $exception) {
            $this->logger->error($exception->getMessage(), [
                'file' => $exception->getFile(),
                'lint' => $exception->getLine(),
            ]);
            throw new ServiceUnavailableHttpException(60, 'Service unavailable, please try later...', $exception);
        }
    }

    /**
     * @Rest\Route("/{cryptobot}/avg/conditions", name="put_avg_conditions", methods={"PUT"})
     * @Rest\View
     *
     * @param Request   $request
     * @param CryptoBot $cryptobot
     *
     * @return array|void
     */
    public function putAvgConditionsAction(Request $request, CryptoBot $cryptobot)
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $form = $this->createForm(TradeConditionCollectionType::class, new TradeConditionCollection($cryptobot), [
            'method' => Request::METHOD_PUT,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            return [
                'form' => $form,
            ];
        }

        /** @var TradeConditionCollection $data */
        $data = $form->getData();

        try {
            $config = $cryptobot->getSymbolConfig($data->symbol);
            if ($config instanceof CryptoTradeConfig) {
                $config->setAvgConditions($data->conditions->toArray());
                $this->cryptoBotService->updateOneTradeLimit($data->cryptobot, $config);
                $this->repository->add($cryptobot, true);
            }
        } catch (\Throwable $exception) {
            $this->logger->error($exception->getMessage(), [
                'file' => $exception->getFile(),
                'lint' => $exception->getLine(),
            ]);
            throw new ServiceUnavailableHttpException(60, 'Service unavailable, please try later...', $exception);
        }
    }

    /**
     * @Rest\Route("/{cryptobot}", name="patch", methods={"PATCH"})
     * @Rest\View(serializerGroups={"cryptobot", "cryptotrade_config", "server", "cryptobot_secured"})
     *
     * @param Request   $request
     * @param CryptoBot $cryptobot
     *
     * @return array|CryptoBot
     */
    public function patchAction(Request $request, CryptoBot $cryptobot)
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $pairsHashBefore = $cryptobot->getPairsHash();
        $form            = $this->createForm(CryptoBotType::class, $cryptobot, [
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

        /** @var CryptoBot $data */
        $data = $form->getData();

        try {
            $this->limitService->checkLimits($data);
        } catch (MaxPairLimitReachedException $exception) {
            throw new BadRequestHttpException($exception->getMessage());
        }

        $this->repository->add($data, true);

        $pairsHashAfter = $data->getPairsHash();
        if ($pairsHashBefore === $pairsHashAfter) {
            if ($data->isRunning()) {
                try {
                    $this->cryptoBotService->setupLimits($data);
                } catch (\Throwable $exception) {
                    $this->logger->error($exception->getMessage(), [
                        'file' => $exception->getFile(),
                        'lint' => $exception->getLine(),
                    ]);
                    throw new ServiceUnavailableHttpException(60, 'Service unavailable, please try later...', $exception);
                }
            }
        } else {
            $cryptobot->setRestartRequired(true);
            $this->repository->add($cryptobot, true);
        }

        return $data;
    }

    /**
     * @Rest\Route("/{cryptobot}", name="delete", methods={"DELETE"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     */
    public function deleteAction(CryptoBot $cryptobot): void
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $this->repository->remove($cryptobot, true);
    }

    /**
     * @Rest\Route("/{cryptobot}/order", name="post_order", methods={"POST"})
     * @Rest\View
     *
     * @param Request   $request
     * @param CryptoBot $cryptobot
     *
     * @return array|void
     */
    public function postOrderAction(Request $request, CryptoBot $cryptobot)
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $form = $this->createForm(ManualOrderType::class, new ManualOrder($cryptobot), [
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

        /** @var ManualOrder $data */
        $data = $form->getData();

        try {
            $this->cryptoBotService->manualOrder($data);
        } catch (\BadMethodCallException $exception) {
            throw new BadRequestHttpException($exception->getMessage(), $exception);
        } catch (\Throwable $exception) {
            $this->logger->error($exception->getMessage(), [
                'file' => $exception->getFile(),
                'lint' => $exception->getLine(),
            ]);
            throw new ServiceUnavailableHttpException(60, 'Service unavailable, please try later...', $exception);
        }
    }

    /**
     * @Rest\Route("/{cryptobot}/order/{symbol}", name="delete_order", methods={"DELETE"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     */
    public function deleteOrderAction(CryptoBot $cryptobot, string $symbol)
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        try {
            $this->cryptoBotService->cancelManualOrder($cryptobot, $symbol);
        } catch (\BadMethodCallException $exception) {
            throw new BadRequestHttpException($exception->getMessage(), $exception);
        } catch (\Throwable $exception) {
            $this->logger->error($exception->getMessage(), [
                'file' => $exception->getFile(),
                'lint' => $exception->getLine(),
            ]);
            throw new ServiceUnavailableHttpException(60, 'Service unavailable, please try later...', $exception);
        }
    }
}
