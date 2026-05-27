<?php
namespace App\Tests\Functional\Bundles\CryptoBotContext\Command;

use App\Tests\RestTestCase;
use Bundles\CryptoBotContext\Command\BotSentimentProcessCommand;
use Symfony\Bundle\FrameworkBundle\Console\Application;
use Symfony\Component\Console\Tester\CommandTester;

class BotSentimentProcessCommandTest extends RestTestCase
{
    public function testShouldAllowToParseNewsAndGetSentiments(): void
    {
        $this->markTestSkipped('Local only');
        $application = new Application(self::$kernel);

        $command       = $application->find(BotSentimentProcessCommand::getDefaultName());
        $commandTester = new CommandTester($command);
        $commandTester->execute([]);

        $commandTester->getDisplay();
    }
}
