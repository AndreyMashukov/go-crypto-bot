<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20240626134409 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_trade_config ADD cfg_position_updated_at DATETIME DEFAULT NULL COMMENT \'(DC2Type:datetime_immutable)\', ADD cfg_quantity DOUBLE PRECISION DEFAULT NULL, ADD cfg_price_today_first DOUBLE PRECISION DEFAULT NULL, ADD cfg_price_today_last DOUBLE PRECISION DEFAULT NULL, ADD cfg_average_price DOUBLE PRECISION DEFAULT NULL');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_trade_config DROP cfg_position_updated_at, DROP cfg_quantity, DROP cfg_price_today_first, DROP cfg_price_today_last, DROP cfg_average_price');
    }
}
