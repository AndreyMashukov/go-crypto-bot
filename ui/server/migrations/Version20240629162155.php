<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20240629162155 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_trade_config ADD cfg_label VARCHAR(20) DEFAULT NULL, ADD cfg_score DOUBLE PRECISION DEFAULT NULL');
        $this->addSql('CREATE INDEX IDX_AC9AB27938A22FD6922FC74 ON rss_article (rsa_expired, rsa_expires_at)');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_trade_config DROP cfg_label, DROP cfg_score');
        $this->addSql('DROP INDEX IDX_AC9AB27938A22FD6922FC74 ON rss_article');
    }
}
