<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\V1;

use App\Service\BackgroundProcessing;
use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Entity\ExchangeSymbol;
use Bundles\CryptoBotContext\Form\CryptoTradeConfigUpdateType;
use Bundles\CryptoBotContext\Form\QuickConfigType;
use Bundles\CryptoBotContext\Model\QuickConfig;
use Bundles\CryptoBotContext\Repository\ExchangeSymbolRepository;
use Bundles\CryptoBotContext\Repository\TradeRepository;
use Bundles\CryptoBotContext\Service\CryptoBotService;
use Bundles\CryptoBotContext\Service\DeployDomain;
use Bundles\CryptoBotContext\Service\TradeListService;
use Bundles\CryptoBotContext\Service\TradeStackService;
use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Exception\MaxPairLimitReachedException;
use Bundles\UserContext\Service\LimitService;
use Doctrine\ORM\EntityManagerInterface;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use GuzzleHttp\Exception\GuzzleException;
use Psr\Log\LoggerInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Exception\AccessDeniedHttpException;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;
use Symfony\Component\HttpKernel\Exception\NotFoundHttpException;
use Symfony\Component\HttpKernel\Exception\ServiceUnavailableHttpException;

/**
 * @Rest\Route("/v1/dashboard", name="v1_dashboard_")
 */
class DashboardController extends AbstractFOSRestController
{
    private CryptoBotService $cryptoBotService;

    private TradeStackService $stackService;

    private EntityManagerInterface $entityManager;

    private TradeRepository $tradeRepository;

    private ExchangeSymbolRepository $symbolRepository;

    private LimitService $limitService;

    private BackgroundProcessing $backgroundProcessing;

    private DeployDomain $deployDomain;

    private LoggerInterface $logger;

    private TradeListService $tradeListService;

    public function __construct(
        CryptoBotService $cryptoBotService,
        TradeStackService $stackService,
        EntityManagerInterface $entityManager,
        TradeRepository $tradeRepository,
        ExchangeSymbolRepository $symbolRepository,
        LimitService $limitService,
        BackgroundProcessing $backgroundProcessing,
        DeployDomain $deployDomain,
        LoggerInterface $logger,
        TradeListService $tradeListService
    ) {
        $this->cryptoBotService     = $cryptoBotService;
        $this->stackService         = $stackService;
        $this->entityManager        = $entityManager;
        $this->tradeRepository      = $tradeRepository;
        $this->symbolRepository     = $symbolRepository;
        $this->limitService         = $limitService;
        $this->backgroundProcessing = $backgroundProcessing;
        $this->deployDomain         = $deployDomain;
        $this->logger               = $logger;
        $this->tradeListService     = $tradeListService;
    }

    /**
     * @Rest\Route("/{cryptobot}/chart", name="chart", methods={"GET"})
     * @Rest\View
     *
     * @param Request   $request
     * @param CryptoBot $cryptobot
     *
     * @return Response
     */
    public function getChartAction(Request $request, CryptoBot $cryptobot): Response
    {
        /** @var User $user */
        $user   = $this->getUser();
        $symbol = $request->get('symbol', '');

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden');
        }

        return new Response(
            $this->cryptoBotService->getChart($cryptobot, $symbol),
            Response::HTTP_OK,
            [
                'Content-Type' => 'application/json',
            ]
        );
    }

    /**
     * @Rest\Route("/{cryptobot}/stack/v2", name="stack_v2", methods={"GET"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     *
     * @return array
     */
    public function getStackV2Action(CryptoBot $cryptobot): array
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden');
        }

        return $this->stackService->getTradeStack($cryptobot);
    }

    /**
     * @Rest\Route("/{cryptobot}/swap/list", name="swap_list", methods={"GET"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     *
     * @return array
     */
    public function getSwapListAction(CryptoBot $cryptobot): array
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden');
        }

        return $this->tradeListService->getSwapActionList($cryptobot);
    }

    /**
     * @Rest\Route("/{cryptobot}/balance", name="balance", methods={"GET"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     *
     * @return array
     */
    public function getBalanceAction(CryptoBot $cryptobot): array
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden');
        }

        try {
            return $this->cryptoBotService->getAccount($cryptobot);
        } catch (\BadMethodCallException|\RuntimeException|\LogicException|GuzzleException $exception) {
            throw new ServiceUnavailableHttpException(60, 'Service unavailable, please try later...', $exception);
        }
    }

    /**
     * @Rest\Route("/{cryptobot}/stack/{sorting}/sort", name="switch_sort", methods={"PUT"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     * @param string    $sorting
     *
     * @return array
     */
    public function putStackSortAction(CryptoBot $cryptobot, string $sorting): array
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden');
        }

        return $this->stackService->setSorting($cryptobot, $sorting);
    }

    /**
     * @Rest\Route("/{cryptobot}/stack/{symbol}/switch", name="switch_symbol", methods={"PUT"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     * @param string    $symbol
     *
     * @return array
     */
    public function putSwitchSymbolAction(CryptoBot $cryptobot, string $symbol): array
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden');
        }

        $result = $this->stackService->switchSymbol($cryptobot, $symbol);
        $config = $cryptobot->getSymbolConfig($result['symbol']);

        if (!$config instanceof CryptoTradeConfig) {
            throw new NotFoundHttpException("Config for {$symbol} is not found.");
        }

        $config->setEnabled($result['isEnabled']);

        try {
            if ($config->isEnabled()) {
                $this->limitService->checkLimits($cryptobot);
            }
        } catch (MaxPairLimitReachedException $exception) {
            throw new BadRequestHttpException($exception->getMessage());
        }

        $this->entityManager->flush();

        return $result;
    }

    /**
     * @Rest\Route("/{cryptobot}/{symbol}/update", name="cryptobot_symbol_update", methods={"PATCH"})
     * @Rest\View
     *
     * @param Request   $request
     * @param CryptoBot $cryptobot
     * @param string    $symbol
     *
     * @return array|CryptoTradeConfig
     */
    public function patchSymbolAction(Request $request, CryptoBot $cryptobot, string $symbol)
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        $config = $cryptobot->getSymbolConfig($symbol);
        if (!$config instanceof CryptoTradeConfig) {
            throw new NotFoundHttpException("Config for {$symbol} is not found.");
        }

        $form = $this->createForm(CryptoTradeConfigUpdateType::class, $config, [
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

        try {
            if ($config->isEnabled()) {
                $this->limitService->checkLimits($cryptobot);
            }
        } catch (MaxPairLimitReachedException $exception) {
            throw new BadRequestHttpException($exception->getMessage());
        }

        $this->cryptoBotService->updateOneTradeLimit($cryptobot, $config);
        $this->entityManager->flush();

        return $config;
    }

    /**
     * @Rest\Route("/{cryptobot}/stack/{symbol}/signal-switch", name="signal_switch_symbol", methods={"PUT"})
     * @Rest\View(serializerGroups={"cryptotrade_config"})
     *
     * @param CryptoBot $cryptobot
     * @param string    $symbol
     *
     * @return CryptoTradeConfig
     */
    public function putSignalSwitchSymbolAction(CryptoBot $cryptobot, string $symbol): CryptoTradeConfig
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden');
        }

        $config = $cryptobot->getSymbolConfig($symbol);
        if (!$config instanceof CryptoTradeConfig) {
            throw new NotFoundHttpException("Config for {$symbol} is not found.");
        }

        $value = !$config->isSignalTrading();

        if ($value && !$user->hasActiveSignalSubscription()) {
            throw new AccessDeniedHttpException("You can't use signal trading without signal subscription. Please contact administration.");
        }

        $config->setSignalTrading($value);
        $this->entityManager->flush();

        return $config;
    }

    /**
     * @Rest\Route("/{cryptobot}/trades", name="trades", methods={"GET"})
     * @Rest\View
     *
     * @return array
     */
    public function getTradesAction(CryptoBot $cryptobot): array
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden');
        }

        try {
            $tradeList = $this->cryptoBotService->getTradeList($cryptobot);
        } catch (\BadMethodCallException $exception) {
            unset($exception);

            return [];
        }

        foreach ($tradeList as $key => $trade) {
            $sellPrecision    = mb_strlen(explode('.', ((string) $trade['sell']))[1] ?? 0);
            $sellQtyPrecision = mb_strlen(explode('.', ((string) $trade['sellQuantity']))[1] ?? 0);

            $tradeList[$key]['buy']         = round($trade['buy'], $sellPrecision);
            $tradeList[$key]['buyQuantity'] = round($trade['buyQuantity'], $sellQtyPrecision);

            $tradeList[$key]['nickname'] = $cryptobot->getUser()->getNickname();
            $tradeList[$key]['profit']   = number_format(round($trade['profit'], 2), 2, '.', '');
        }

        return $tradeList;
    }

    /**
     * @Rest\Route("/{cryptobot}/symbol/available", name="symbol_available", methods={"GET"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     *
     * @return array
     */
    public function getSymbolAvailableAction(CryptoBot $cryptobot): array
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden');
        }

        return $this->symbolRepository->getAvailableSymbols($cryptobot);
    }

    /**
     * @Rest\Route("/{cryptobot}/profit", name="profit", methods={"GET"})
     * @Rest\View
     *
     * @param CryptoBot $cryptobot
     * @param Request   $request
     *
     * @return array
     */
    public function getProfitAction(CryptoBot $cryptobot, Request $request): array
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden');
        }

        $period = $request->get('period', 'day');

        if (!\in_array($period, ['day', 'week', 'month'], true)) {
            throw new BadRequestHttpException('Invalid period given.');
        }

        return $this->tradeRepository->getProfitByPeriod($cryptobot, $period);
    }

    /**
     * @Rest\Route("/{cryptobot}/positions", name="positions", methods={"GET"})
     * @Rest\View
     *
     * @return array
     */
    public function getPositionsAction(CryptoBot $cryptobot): array
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden');
        }

        return $this->tradeListService->getBotPositionList($cryptobot);
    }

    /**
     * @Rest\Route("/{cryptobot}/quick/symbol", name="quick_symbol", methods={"POST"})
     * @Rest\View
     *
     * @param Request   $request
     * @param CryptoBot $cryptobot
     *
     * @return array|void
     */
    public function postQuickSymbolAction(Request $request, CryptoBot $cryptobot)
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$cryptobot->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        if (!$cryptobot->isRunning()) {
            throw new BadRequestHttpException('Bot is not running.');
        }

        $form = $this->createForm(QuickConfigType::class, new QuickConfig($cryptobot), [
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

        /** @var QuickConfig $data */
        $data = $form->getData();

        try {
            $this->limitService->checkLimits($cryptobot);
        } catch (MaxPairLimitReachedException $exception) {
            throw new BadRequestHttpException($exception->getMessage());
        }

        $exchangeSymbol = $data->getExchangeSymbol();
        $addedSymbol    = $exchangeSymbol->getSymbol();
        $exchange       = $exchangeSymbol->getExchange();

        if ($exchange !== $cryptobot->getProvider()) {
            throw new BadRequestHttpException("Wrong provider given: {$exchange}");
        }

        if (!$exchangeSymbol->isEnabled()) {
            throw new BadRequestHttpException("Symbol {$addedSymbol} is not enabled for {$exchange}");
        }

        /** @var ExchangeSymbol[] $availableSymbols */
        $availableSymbols  = $this->symbolRepository->getAvailableSymbols($cryptobot);
        $isAllowedToAdd    = false;
        foreach ($availableSymbols as $symbol) {
            if ($addedSymbol === $symbol->getSymbol()) {
                $isAllowedToAdd = true;
            }
        }

        if (!$isAllowedToAdd) {
            throw new BadRequestHttpException("Is not allowed to add Symbol: {$addedSymbol}");
        }

        $cryptoConfig = new CryptoTradeConfig();
        $cryptoConfig->setSymbol($addedSymbol)
            ->setAvgConditions([])
            ->setBuyConditions([
                [
                    'type'     => 'or',
                    'value'    => null,
                    'symbol'   => null,
                    'children' => [
                        [
                            'symbol'    => $addedSymbol,
                            'parameter' => 'has_signal',
                            'condition' => 'eq',
                            'value'     => 'true',
                            'type'      => 'and',
                            'children'  => [],
                        ],
                        [
                            'symbol'    => $addedSymbol,
                            'parameter' => 'daily_percent',
                            'condition' => 'lte',
                            'value'     => '-5.00',
                            'type'      => 'and',
                            'children'  => [],
                        ],
                    ],
                ],
            ])
            ->setSellConditions([])
            ->setUsdtLimit(50)
            ->setEnabled(false)
            ->setExtraChargeOptions([])
            ->setProfitOptions([
                [
                    'index'           => 0,
                    'optionUnit'      => 'h',
                    'optionValue'     => 1,
                    'optionPercent'   => 5.00,
                    'isTriggerOption' => true,
                ],
            ])
            ->setSignalTrading(true)
            ->setCryptobot($cryptobot);
        $cryptobot->addCryptoTradeConfig($cryptoConfig);
        $this->entityManager->persist($cryptoConfig);

        if (!$data->isRestartBot()) {
            $cryptobot->setRestartRequired(true);
        }

        $this->entityManager->flush();

        if ($data->isRestartBot()) {
            // Redeploy bot
            $this->backgroundProcessing->addTask(function () use ($cryptobot) {
                $this->deployDomain->doDeploy($cryptobot);
            });
        } else {
            try {
                $this->cryptoBotService->setupLimits($cryptobot);
            } catch (\Exception $exception) {
                $this->logger->error($exception->getMessage(), [
                    'file' => $exception->getFile(),
                    'lint' => $exception->getLine(),
                ]);
                throw new ServiceUnavailableHttpException(60, 'Service unavailable, please try later...', $exception);
            }
        }
    }
}
