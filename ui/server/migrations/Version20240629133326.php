<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20240629133326 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TABLE rss_article (rsa_id INT AUTO_INCREMENT NOT NULL, rsa_url VARCHAR(255) NOT NULL, rsa_url_crc_32 BIGINT NOT NULL, rsa_expires_at DATETIME NOT NULL COMMENT \'(DC2Type:datetime_immutable)\', rsa_label VARCHAR(20) NOT NULL, rsa_score DOUBLE PRECISION NOT NULL, rsa_expired TINYINT(1) DEFAULT 0 NOT NULL, rsa_coins JSON NOT NULL, rsa_created_at DATETIME NOT NULL COMMENT \'(DC2Type:datetime_immutable)\', INDEX IDX_AC9AB275726BCA66664554E (rsa_url, rsa_url_crc_32), PRIMARY KEY(rsa_id)) DEFAULT CHARACTER SET utf8mb4 COLLATE `utf8mb4_unicode_ci` ENGINE = InnoDB');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('DROP TABLE rss_article');
    }
}
