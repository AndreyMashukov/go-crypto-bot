<?php
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
    public function __construct(private readonly DeployService $deployService, private readonly CryptoBotService $cryptoBotService, private readonly CryptoBotRepository $cryptoBotRepository, private readonly ServerRepository $serverRepository, private readonly EventDispatcherInterface $eventDispatcher, private readonly LoggerInterface $logger, private readonly EntityManagerInterface $entityManager)
    {
    }

    /**
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

            if ($cryptobot->getServer() instanceof Server) {
                $this->doStop($cryptobot, $cryptobot->getServer(), '', false);
            }

            $server = $this->serverRepository->getAvailableServer($cryptobot);

            if (!$server instanceof Server) {
                throw new ServersIsOutOfStockException('Server is not available.');
            }

            if (!$this->deployService->deploy($cryptobot, $server)) {
                $port        = $cryptobot->getPort();
                $containerId = $cryptobot->getContainerId();
                $errorMessage = $cryptobot->getErrorMessage();

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

            throw new \DomainException("Couldn't deploy the bot, please try later", $exception->getCode(), $exception);
        }
    }

    public function doStop(CryptoBot $cryptobot, Server $server, string $reason, bool $save): void
    {
        $cryptobot->setErrorMessage(null);

        if ($reason !== '' && $reason !== '0') {
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
