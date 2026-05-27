<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Command;

use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Event\BudgetPurchaseEvent;
use Bundles\UserContext\Repository\UserRepository;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Input\InputOption;
use Symfony\Component\Console\Output\OutputInterface;
use Symfony\Component\Console\Style\SymfonyStyle;
use Symfony\Component\EventDispatcher\EventDispatcherInterface;

class UserBudgetComplementaryCommand extends Command
{
    protected static $defaultName = 'user:budget:complementary';

    protected static $defaultDescription = 'Complementary budget purchase';

    private EventDispatcherInterface $eventDispatcher;

    private EntityManagerInterface $entityManager;

    private UserRepository $userRepository;

    public function __construct(
        EventDispatcherInterface $eventDispatcher,
        EntityManagerInterface $entityManager,
        UserRepository $userRepository
    ) {
        parent::__construct(self::$defaultName);

        $this->eventDispatcher = $eventDispatcher;
        $this->entityManager   = $entityManager;
        $this->userRepository  = $userRepository;
    }

    protected function configure(): void
    {
        $this
            ->addOption('amount', 'a', InputOption::VALUE_REQUIRED, 'Budget')
            ->addOption('user', 'u', InputOption::VALUE_REQUIRED, 'User')
        ;
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        $io = new SymfonyStyle($input, $output);

        $amount = $input->getOption('amount');

        if (!$amount) {
            $io->note('Amount should be greater then 0');

            return Command::SUCCESS;
        }

        $userId = $input->getOption('user');
        $user   = $this->userRepository->find($userId ?: 0);
        if (!$user instanceof User) {
            $io->note('User is not found.');

            return Command::SUCCESS;
        }

        $this->eventDispatcher->dispatch(new BudgetPurchaseEvent($user, $amount));
        $this->entityManager->flush();

        return Command::SUCCESS;
    }
}
