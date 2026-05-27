<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20240318141831 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_bot ADD ctb_dedicated_server INT DEFAULT NULL');
        $this->addSql('ALTER TABLE crypto_bot ADD CONSTRAINT FK_BE990B7ABDD959AE FOREIGN KEY (ctb_dedicated_server) REFERENCES server (srv_id)');
        $this->addSql('CREATE INDEX IDX_BE990B7ABDD959AE ON crypto_bot (ctb_dedicated_server)');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_bot DROP FOREIGN KEY FK_BE990B7ABDD959AE');
        $this->addSql('DROP INDEX IDX_BE990B7ABDD959AE ON crypto_bot');
        $this->addSql('ALTER TABLE crypto_bot DROP ctb_dedicated_server');
    }
}
