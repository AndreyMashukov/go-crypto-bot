<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20240603153522 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_trade_config ADD signal_config_percent_filter DOUBLE PRECISION DEFAULT \'1\' NOT NULL, ADD signal_config_rating_filter TINYINT(1) DEFAULT 0 NOT NULL, ADD signal_config_avg_buy_filter TINYINT(1) DEFAULT 0 NOT NULL, ADD signal_config_avg_sell_filter TINYINT(1) DEFAULT 0 NOT NULL, ADD signal_config_avg_buy_correction TINYINT(1) DEFAULT 1 NOT NULL, ADD signal_config_avg_sell_correction TINYINT(1) DEFAULT 0 NOT NULL');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_trade_config DROP signal_config_percent_filter, DROP signal_config_rating_filter, DROP signal_config_avg_buy_filter, DROP signal_config_avg_sell_filter, DROP signal_config_avg_buy_correction, DROP signal_config_avg_sell_correction');
    }
}
