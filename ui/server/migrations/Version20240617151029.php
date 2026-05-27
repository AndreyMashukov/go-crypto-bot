<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20240617151029 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE payment DROP FOREIGN KEY FK_6D28840D88C7B305');
        $this->addSql('ALTER TABLE user DROP FOREIGN KEY FK_8D93D649A3C664D3');
        $this->addSql('DROP TABLE subscription');
        $this->addSql('DROP INDEX IDX_6D28840D88C7B305 ON payment');
        $this->addSql('ALTER TABLE payment DROP pmt_subscription');
        $this->addSql('ALTER TABLE user DROP FOREIGN KEY FK_8D93D6491E3D6496');
        $this->addSql('DROP INDEX IDX_8D93D6491E3D6496 ON user');
        $this->addSql('DROP INDEX IDX_8D93D649A3C664D3 ON user');
        $this->addSql('ALTER TABLE user DROP subscription_payment, DROP subscription');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TABLE subscription (sub_id INT AUTO_INCREMENT NOT NULL, sub_name VARCHAR(255) CHARACTER SET utf8mb4 NOT NULL COLLATE `utf8mb4_unicode_ci`, sub_budget_limit DOUBLE PRECISION NOT NULL, sub_extra_budget_limit DOUBLE PRECISION NOT NULL, sub_price DOUBLE PRECISION NOT NULL, sub_period_days INT NOT NULL, sub_max_symbols INT NOT NULL, sub_grace_period_days INT NOT NULL, sub_price_featured DOUBLE PRECISION DEFAULT NULL, PRIMARY KEY(sub_id)) DEFAULT CHARACTER SET utf8mb4 COLLATE `utf8mb4_unicode_ci` ENGINE = InnoDB COMMENT = \'\' ');
        $this->addSql('ALTER TABLE payment ADD pmt_subscription INT DEFAULT NULL');
        $this->addSql('ALTER TABLE payment ADD CONSTRAINT FK_6D28840D88C7B305 FOREIGN KEY (pmt_subscription) REFERENCES subscription (sub_id)');
        $this->addSql('CREATE INDEX IDX_6D28840D88C7B305 ON payment (pmt_subscription)');
        $this->addSql('ALTER TABLE user ADD subscription_payment INT DEFAULT NULL, ADD subscription INT DEFAULT NULL');
        $this->addSql('ALTER TABLE user ADD CONSTRAINT FK_8D93D649A3C664D3 FOREIGN KEY (subscription) REFERENCES subscription (sub_id)');
        $this->addSql('ALTER TABLE user ADD CONSTRAINT FK_8D93D6491E3D6496 FOREIGN KEY (subscription_payment) REFERENCES payment (pmt_id)');
        $this->addSql('CREATE INDEX IDX_8D93D6491E3D6496 ON user (subscription_payment)');
        $this->addSql('CREATE INDEX IDX_8D93D649A3C664D3 ON user (subscription)');
    }
}
