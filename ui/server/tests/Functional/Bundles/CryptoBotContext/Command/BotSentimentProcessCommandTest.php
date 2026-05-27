<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Functional\Bundles\CryptoBotContext\Command;

use App\Tests\RestTestCase;
use Bundles\CryptoBotContext\Command\BotSentimentProcessCommand;
use Symfony\Bundle\FrameworkBundle\Console\Application;
use Symfony\Component\Console\Tester\CommandTester;

/**
 * @group functional
 */
class BotSentimentProcessCommandTest extends RestTestCase
{
    /**
     * Should allow to parse news and get sentiments.
     */
    public function testShouldAllowToParseNewsAndGetSentiments(): void
    {
        //return;
        $this->markTestSkipped('Local only');
        $application = new Application(self::$kernel);

        $command       = $application->find(BotSentimentProcessCommand::getDefaultName());
        $commandTester = new CommandTester($command);
        $commandTester->execute([]);

        // the output of the command in the console
        $output = $commandTester->getDisplay();
    }
}
