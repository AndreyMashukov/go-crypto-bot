<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20231225160222 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_trade_config ADD cfg_min_price_minutes_period INT DEFAULT 200 NOT NULL, ADD cfg_frame_interval VARCHAR(5) DEFAULT \'2h\' NOT NULL, ADD cfg_frame_period INT DEFAULT 20 NOT NULL, ADD cfg_buy_price_history_check_interval VARCHAR(5) DEFAULT \'1d\' NOT NULL, ADD cfg_buy_price_history_check_period INT DEFAULT 14 NOT NULL');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_trade_config DROP cfg_min_price_minutes_period, DROP cfg_frame_interval, DROP cfg_frame_period, DROP cfg_buy_price_history_check_interval, DROP cfg_buy_price_history_check_period');
    }
}
