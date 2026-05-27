<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20231205110743 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE user ADD subscription_payment INT DEFAULT NULL');
        $this->addSql('ALTER TABLE user ADD CONSTRAINT FK_8D93D6491E3D6496 FOREIGN KEY (subscription_payment) REFERENCES payment (pmt_id)');
        $this->addSql('CREATE INDEX IDX_8D93D6491E3D6496 ON user (subscription_payment)');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE user DROP FOREIGN KEY FK_8D93D6491E3D6496');
        $this->addSql('DROP INDEX IDX_8D93D6491E3D6496 ON user');
        $this->addSql('ALTER TABLE user DROP subscription_payment');
    }
}
