<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20240525111701 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        // todo: Add support for ENUM, now change to `exchange CHAR(10) NOT NULL` during DEV activities.
        $this->addSql('CREATE TABLE exchange_symbol (id INT AUTO_INCREMENT NOT NULL, symbol VARCHAR(10) NOT NULL, exchange ENUM(\'binance\',\'bybit\') NOT NULL, enabled TINYINT(1) NOT NULL, PRIMARY KEY(id)) DEFAULT CHARACTER SET utf8mb4 COLLATE `utf8mb4_unicode_ci` ENGINE = InnoDB');
        $symbols = [
            'NEOUSDT',
            'PERPUSDT',
            'ETHUSDT',
            'SOLUSDT',
            'BTCUSDT',
            'LTCUSDT',
            'XRPUSDT',
            'BNBUSDT',
            'TRXUSDT',
            'AVAXUSDT',
            'ADAUSDT',
            'DOGEUSDT',
            'BCHUSDT',
            'LINKUSDT',
            'MATICUSDT',
            'DOTUSDT',
            'UNIUSDT',
            'ETCUSDT',
            'XLMUSDT',
            'ATOMUSDT',
            'NEARUSDT',
            'ZECUSDT',
            'SHIBUSDT',
        ];
        foreach ($symbols as $symbol) {
            foreach ([CryptoBot::PROVIDER_BINANCE, CryptoBot::PROVIDER_BYBIT] as $exchange) {
                $isEnabled = 1;
                if ('ZECUSDT' === $symbol && CryptoBot::PROVIDER_BYBIT === $exchange) {
                    $isEnabled = 0;
                }
                $this->addSql("INSERT INTO exchange_symbol SET symbol = '{$symbol}', exchange = '{$exchange}', enabled = {$isEnabled}");
            }
        }
        $this->addSql('ALTER TABLE crypto_bot ADD ctb_restart_required TINYINT(1) DEFAULT 0 NOT NULL');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('DROP TABLE exchange_symbol');
        $this->addSql('ALTER TABLE crypto_bot DROP COLUMN ctb_restart_required');
    }
}
