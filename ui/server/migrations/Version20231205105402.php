<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20231205105402 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TABLE payment (pmt_id INT AUTO_INCREMENT NOT NULL, pmt_subscription INT NOT NULL, pmt_user INT NOT NULL, pmt_amount DOUBLE PRECISION NOT NULL, pmt_status VARCHAR(255) NOT NULL, pmt_currency VARCHAR(255) NOT NULL, pmt_description VARCHAR(255) NOT NULL, pmt_order_id CHAR(36) NOT NULL COMMENT \'(DC2Type:uuid)\', pmt_email VARCHAR(255) NOT NULL, pmt_created_at DATETIME NOT NULL COMMENT \'(DC2Type:datetime_immutable)\', pmt_track_id INT DEFAULT NULL, pmt_payment_link VARCHAR(255) DEFAULT NULL, pmt_expires_at DATETIME DEFAULT NULL COMMENT \'(DC2Type:datetime_immutable)\', pmt_completed_at DATETIME DEFAULT NULL COMMENT \'(DC2Type:datetime_immutable)\', INDEX IDX_6D28840D88C7B305 (pmt_subscription), INDEX IDX_6D28840DA5347D44 (pmt_user), PRIMARY KEY(pmt_id)) DEFAULT CHARACTER SET utf8mb4 COLLATE `utf8mb4_unicode_ci` ENGINE = InnoDB');
        $this->addSql('CREATE TABLE subscription (sub_id INT AUTO_INCREMENT NOT NULL, sub_name VARCHAR(255) NOT NULL, sub_budget_limit DOUBLE PRECISION NOT NULL, sub_extra_budget_limit DOUBLE PRECISION NOT NULL, sub_price DOUBLE PRECISION NOT NULL, sub_period_days INT NOT NULL, sub_max_symbols INT NOT NULL, sub_grace_period_days INT NOT NULL, PRIMARY KEY(sub_id)) DEFAULT CHARACTER SET utf8mb4 COLLATE `utf8mb4_unicode_ci` ENGINE = InnoDB');
        $this->addSql('ALTER TABLE payment ADD CONSTRAINT FK_6D28840D88C7B305 FOREIGN KEY (pmt_subscription) REFERENCES subscription (sub_id)');
        $this->addSql('ALTER TABLE payment ADD CONSTRAINT FK_6D28840DA5347D44 FOREIGN KEY (pmt_user) REFERENCES user (id)');
        $this->addSql('ALTER TABLE user CHANGE nickname nickname VARCHAR(70) NOT NULL');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE payment DROP FOREIGN KEY FK_6D28840D88C7B305');
        $this->addSql('ALTER TABLE payment DROP FOREIGN KEY FK_6D28840DA5347D44');
        $this->addSql('DROP TABLE payment');
        $this->addSql('DROP TABLE subscription');
        $this->addSql('ALTER TABLE user CHANGE nickname nickname VARCHAR(70) DEFAULT NULL');
    }
}
