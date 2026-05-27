<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20231201082623 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TABLE crypto_bot (ctb_id INT AUTO_INCREMENT NOT NULL, ctb_user INT NOT NULL, ctb_uuid CHAR(36) NOT NULL COMMENT \'(DC2Type:uuid)\', ctb_provider VARCHAR(50) NOT NULL, ctb_api_key VARCHAR(255) DEFAULT NULL, ctb_api_secret VARCHAR(255) DEFAULT NULL, ctb_ip_address VARCHAR(255) DEFAULT NULL, ctb_port VARCHAR(255) DEFAULT NULL, ctb_container_id VARCHAR(255) DEFAULT NULL, ctb_status VARCHAR(255) NOT NULL, INDEX IDX_BE990B7A70C97317 (ctb_user), PRIMARY KEY(ctb_id)) DEFAULT CHARACTER SET utf8mb4 COLLATE `utf8mb4_unicode_ci` ENGINE = InnoDB');
        $this->addSql('ALTER TABLE crypto_bot ADD CONSTRAINT FK_BE990B7A70C97317 FOREIGN KEY (ctb_user) REFERENCES user (id)');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_bot DROP FOREIGN KEY FK_BE990B7A70C97317');
        $this->addSql('DROP TABLE crypto_bot');
    }
}
