<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20231201090833 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TABLE crypto_trade_config (cfg_id INT AUTO_INCREMENT NOT NULL, cfg_cryptobot INT NOT NULL, cfg_symbol VARCHAR(255) NOT NULL, cfg_usdt_limit DOUBLE PRECISION NOT NULL, cfg_min_profit_percent DOUBLE PRECISION NOT NULL, cfg_is_enabled TINYINT(1) NOT NULL, cfg_buy_on_fall_percent DOUBLE PRECISION NOT NULL, cfg_usdt_extra_budget DOUBLE PRECISION NOT NULL, INDEX IDX_EAEFE2D322D2A868 (cfg_cryptobot), PRIMARY KEY(cfg_id)) DEFAULT CHARACTER SET utf8mb4 COLLATE `utf8mb4_unicode_ci` ENGINE = InnoDB');
        $this->addSql('ALTER TABLE crypto_trade_config ADD CONSTRAINT FK_EAEFE2D322D2A868 FOREIGN KEY (cfg_cryptobot) REFERENCES crypto_bot (ctb_id)');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_trade_config DROP FOREIGN KEY FK_EAEFE2D322D2A868');
        $this->addSql('DROP TABLE crypto_trade_config');
    }
}
