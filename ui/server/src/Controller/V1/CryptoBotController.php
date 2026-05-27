<?php
namespace App\Controller\V1;

use Bundles\CryptoBotContext\Entity\Server;
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
use Bundles\UserContext\Exception\MaxPairLimitReachedException;
use Bundles\UserContext\Service\LimitService;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Psr\Log\LoggerInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Exception\AccessDeniedHttpException;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;
use Symfony\Component\HttpKernel\Exception\ServiceUnavailableHttpException;

class CryptoBotController extends AbstractFOSRestController
{
    public const AVAILABLE_PROVIDERS = [
        CryptoBot::PROVIDER_BINANCE => true,
        CryptoBot::PROVIDER_BYBIT   => true,
    ];

    public function __construct(private readonly CryptoBotRepository $repository, private readonly DeployDomain $deployDomain, private readonly CryptoBotService $cryptoBotService, private readonly ServerRepository $serverRepository, private readonly LoggerInterface $logger, private readonly ExchangeSymbolRepository $symbolRepository, private readonly LimitService $limitService, private readonly CommissionService $commissionService, private readonly AlertService $alertService)
    {
    }

    public function getList(): array
    {
        $user = $this->getUser();

        return $this->repository->findBy([
            'user' => $user,
        ]);
    }

    public function getListExtended(): array
    {
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

    public function getSymbolList(CryptoBot $cryptobot): array
    {
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        return $this->symbolRepository->findBy([
            'exchange' => $cryptobot->getProvider(),
            'enabled'  => true,
        ], ['symbol' => 'ASC']);
    }

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

    public function getBotServerList(CryptoBot $cryptobot): array
    {
        $list = $cryptobot->hasDedicatedServer() ? [
            $cryptobot->getDedicated(),
        ] : $this->serverRepository->findAll();

        $stats = $this->serverRepository->getServerStats();

        return [
            'list'    => $list,
            'servers' => $stats,
        ];
    }

    public function getAction(CryptoBot $cryptobot): CryptoBot
    {
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        return $cryptobot;
    }

    public function putDeploy(CryptoBot $cryptobot): void
    {
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

    public function putStop(CryptoBot $cryptobot): void
    {
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        if (!$cryptobot->getServer() instanceof Server) {
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

    public function putSyncConfig(CryptoBot $cryptobot): void
    {
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $this->cryptoBotService->syncConfig($cryptobot);
        $this->repository->add($cryptobot, true);
    }

    /**
     * @return array|CryptoBot
     */
    public function post(Request $request)
    {
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
     * @return mixed[]|null
     */
    public function putMultiCharge(Request $request, CryptoBot $cryptobot)
    {
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
        return null;
    }

    /**
     * @return mixed[]|null
     */
    public function putMultiProfit(Request $request, CryptoBot $cryptobot)
    {
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
        return null;
    }

    /**
     * @return mixed[]|null
     */
    public function putBuyConditions(Request $request, CryptoBot $cryptobot)
    {
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
        return null;
    }

    /**
     * @return mixed[]|null
     */
    public function putSellConditions(Request $request, CryptoBot $cryptobot)
    {
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
        return null;
    }

    /**
     * @return mixed[]|null
     */
    public function putAvgConditions(Request $request, CryptoBot $cryptobot)
    {
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
        return null;
    }

    /**
     * @return array|CryptoBot
     */
    public function patch(Request $request, CryptoBot $cryptobot)
    {
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

        $data = $form->getData();

        try {
            $this->limitService->checkLimits($data);
        } catch (MaxPairLimitReachedException $exception) {
            throw new BadRequestHttpException($exception->getMessage(), $exception);
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

    public function delete(CryptoBot $cryptobot): void
    {
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $this->repository->remove($cryptobot, true);
    }

    /**
     * @return mixed[]|null
     */
    public function postOrder(Request $request, CryptoBot $cryptobot)
    {
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
        return null;
    }

    public function deleteOrder(CryptoBot $cryptobot, string $symbol)
    {
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
