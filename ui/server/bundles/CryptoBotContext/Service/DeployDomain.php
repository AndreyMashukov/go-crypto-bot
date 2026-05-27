<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\Server;
use Bundles\CryptoBotContext\Event\StopBotEvent;
use Bundles\CryptoBotContext\Exception\ServersIsOutOfStockException;
use Bundles\CryptoBotContext\Repository\CryptoBotRepository;
use Bundles\CryptoBotContext\Repository\ServerRepository;
use Doctrine\ORM\EntityManagerInterface;
use Psr\Log\LoggerInterface;
use Symfony\Component\EventDispatcher\EventDispatcherInterface;

class DeployDomain
{
    private DeployService $deployService;

    private CryptoBotService $cryptoBotService;

    private CryptoBotRepository $cryptoBotRepository;

    private ServerRepository $serverRepository;

    private EventDispatcherInterface $eventDispatcher;

    private LoggerInterface $logger;

    private EntityManagerInterface $entityManager;

    public function __construct(
        DeployService $deployService,
        CryptoBotService $cryptoBotService,
        CryptoBotRepository $cryptoBotRepository,
        ServerRepository $serverRepository,
        EventDispatcherInterface $eventDispatcher,
        LoggerInterface $logger,
        EntityManagerInterface $entityManager
    ) {
        $this->deployService             = $deployService;
        $this->cryptoBotService          = $cryptoBotService;
        $this->cryptoBotRepository       = $cryptoBotRepository;
        $this->serverRepository          = $serverRepository;
        $this->eventDispatcher           = $eventDispatcher;
        $this->logger                    = $logger;
        $this->entityManager             = $entityManager;
    }

    /**
     * @param CryptoBot $cryptobot
     *
     * @throws ServersIsOutOfStockException
     * @throws \DomainException
     */
    public function doDeploy(CryptoBot $cryptobot): void
    {
        try {
            $cryptobot->setErrorMessage(null);
            $cryptobot->setStatus(CryptoBot::STATUS_DEPLOY);
            $this->cryptoBotRepository->add($cryptobot, true);

            // todo: rent server (book)
            if ($cryptobot->getServer()) {
                $this->doStop($cryptobot, $cryptobot->getServer(), '', false);
            }

            $server = $this->serverRepository->getAvailableServer($cryptobot);

            if (!$server instanceof Server) {
                throw new ServersIsOutOfStockException('Server is not available.');
            }

            if (!$this->deployService->deploy($cryptobot, $server)) {
                $port        = $cryptobot->getPort();
                $containerId = $cryptobot->getContainerId();
                // Port and container ID is set after deploy, and will be rewritten after refresh.
                $errorMessage = $cryptobot->getErrorMessage();

                // Refresh, because we can get webhook
                $this->entityManager->refresh($cryptobot);
                $cryptobot->setPort($port);
                $cryptobot->setContainerId($containerId);

                $remoteErrorMessage = $cryptobot->getErrorMessage();
                if ($remoteErrorMessage) {
                    $errorMessage = $remoteErrorMessage;
                }

                if (!$errorMessage) {
                    $errorMessage = "Couldn't deploy the bot, please try later";
                }

                $this->doStop($cryptobot, $server, $errorMessage, false);
                $cryptobot->setStatus(CryptoBot::STATUS_STOPPED);
                $cryptobot->setErrorMessage($errorMessage);

                $currentServer = $cryptobot->getServer();

                if ($currentServer instanceof Server && !$cryptobot->isDedicated($currentServer)) {
                    $cryptobot->setServer(null);
                }

                $this->cryptoBotRepository->add($cryptobot, true);

                throw new \DomainException($errorMessage);
            }

            $cryptobot->setServer($server);
            $this->cryptoBotService->setupLimits($cryptobot);
            $cryptobot->setRestartRequired(false);
            $this->cryptoBotRepository->add($cryptobot, true);
        } catch (\Throwable $exception) {
            $this->logger->error($exception->getMessage(), [
                'line' => $exception->getLine(),
                'file' => $exception->getFile(),
            ]);

            if ($exception instanceof \DomainException) {
                throw $exception;
            }

            $cryptobot->setStatus(CryptoBot::STATUS_ERROR);
            $this->cryptoBotRepository->add($cryptobot, true);

            throw new \DomainException("Couldn't deploy the bot, please try later");
        }
    }

    public function doStop(CryptoBot $cryptobot, Server $server, string $reason, bool $save): void
    {
        $cryptobot->setErrorMessage(null);

        if ($reason) {
            $cryptobot->setErrorMessage($reason);
        }

        try {
            $this->deployService->stop($cryptobot, $server);
        } catch (\LogicException $exception) {
            unset($exception);
            $cryptobot->setStatus(CryptoBot::STATUS_STOPPED);
        }

        $this->eventDispatcher->dispatch(new StopBotEvent($cryptobot));

        if ($save) {
            $this->cryptoBotRepository->add($cryptobot, true);
        }
    }
}
