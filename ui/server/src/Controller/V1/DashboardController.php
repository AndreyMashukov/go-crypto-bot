<?php
namespace App\Controller\V1;

use App\Service\BackgroundProcessing;
use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Form\CryptoTradeConfigUpdateType;
use Bundles\CryptoBotContext\Form\QuickConfigType;
use Bundles\CryptoBotContext\Model\QuickConfig;
use Bundles\CryptoBotContext\Repository\ExchangeSymbolRepository;
use Bundles\CryptoBotContext\Repository\TradeRepository;
use Bundles\CryptoBotContext\Service\CryptoBotService;
use Bundles\CryptoBotContext\Service\DeployDomain;
use Bundles\CryptoBotContext\Service\TradeListService;
use Bundles\CryptoBotContext\Service\TradeStackService;
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

class DashboardController extends AbstractFOSRestController
{
    public function __construct(private readonly CryptoBotService $cryptoBotService, private readonly TradeStackService $stackService, private readonly EntityManagerInterface $entityManager, private readonly TradeRepository $tradeRepository, private readonly ExchangeSymbolRepository $symbolRepository, private readonly LimitService $limitService, private readonly BackgroundProcessing $backgroundProcessing, private readonly DeployDomain $deployDomain, private readonly LoggerInterface $logger, private readonly TradeListService $tradeListService)
    {
    }

    /**
     * @Rest\Route("/{cryptobot}/chart", name="chart", methods={"GET"})
     * @Rest\View
     *
     *
     */
    public function getChart(Request $request, CryptoBot $cryptobot): Response
    {
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
     *
     */
    public function getStackV2(CryptoBot $cryptobot): array
    {
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
     *
     */
    public function getSwapList(CryptoBot $cryptobot): array
    {
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
     *
     */
    public function getBalance(CryptoBot $cryptobot): array
    {
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
     *
     */
    public function putStackSort(CryptoBot $cryptobot, string $sorting): array
    {
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
     *
     */
    public function putSwitchSymbol(CryptoBot $cryptobot, string $symbol): array
    {
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
            throw new BadRequestHttpException($exception->getMessage(), $exception);
        }

        $this->entityManager->flush();

        return $result;
    }

    /**
     * @Rest\Route("/{cryptobot}/{symbol}/update", name="cryptobot_symbol_update", methods={"PATCH"})
     * @Rest\View
     *
     *
     * @return array|CryptoTradeConfig
     */
    public function patchSymbol(Request $request, CryptoBot $cryptobot, string $symbol)
    {
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
            throw new BadRequestHttpException($exception->getMessage(), $exception);
        }

        $this->cryptoBotService->updateOneTradeLimit($cryptobot, $config);
        $this->entityManager->flush();

        return $config;
    }

    /**
     * @Rest\Route("/{cryptobot}/stack/{symbol}/signal-switch", name="signal_switch_symbol", methods={"PUT"})
     * @Rest\View(serializerGroups={"cryptotrade_config"})
     *
     *
     */
    public function putSignalSwitchSymbol(CryptoBot $cryptobot, string $symbol): CryptoTradeConfig
    {
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
     */
    public function getTrades(CryptoBot $cryptobot): array
    {
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
            $sellPrecision    = mb_strlen((string) (explode('.', ((string) $trade['sell']))[1] ?? 0));
            $sellQtyPrecision = mb_strlen((string) (explode('.', ((string) $trade['sellQuantity']))[1] ?? 0));

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
     *
     */
    public function getSymbolAvailable(CryptoBot $cryptobot): array
    {
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
     *
     */
    public function getProfit(CryptoBot $cryptobot, Request $request): array
    {
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
     */
    public function getPositions(CryptoBot $cryptobot): array
    {
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
     *
     * @return mixed[]|null
     */
    public function postQuickSymbol(Request $request, CryptoBot $cryptobot)
    {
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

        $data = $form->getData();

        try {
            $this->limitService->checkLimits($cryptobot);
        } catch (MaxPairLimitReachedException $exception) {
            throw new BadRequestHttpException($exception->getMessage(), $exception);
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
        return null;
    }
}
