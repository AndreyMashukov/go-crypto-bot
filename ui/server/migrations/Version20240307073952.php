<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20240307073952 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_trade_config ADD cfg_profit_options JSON NOT NULL');
        $this->addSql('UPDATE crypto_trade_config SET cfg_profit_options = JSON_ARRAY() WHERE crypto_trade_config.cfg_id > 0');
        $this->addSql('UPDATE crypto_trade_config SET cfg_profit_options = CAST(CONCAT(\'[{"index": 0, "isTriggerOption": true, "optionUnit": "h", "optionValue": 1, "optionPercent": \', cfg_min_profit_percent, \'}]\') as JSON) WHERE cfg_id > 0;');
        $this->addSql('ALTER TABLE crypto_trade_config DROP COLUMN cfg_min_profit_percent');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_trade_config DROP cfg_profit_options');
    }
}
