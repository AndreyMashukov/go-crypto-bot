<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20240603023754 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TABLE promo_code (pcd_id INT AUTO_INCREMENT NOT NULL, pcd_partner INT NOT NULL, pcd_complimentary_budget DOUBLE PRECISION NOT NULL, pcd_complimentary_signals_days SMALLINT NOT NULL, pcd_code VARCHAR(20) NOT NULL, pcd_max_activation_limit INT NOT NULL, pcd_activation_count INT NOT NULL, pcd_partner_fee_percent DOUBLE PRECISION NOT NULL, pcd_active TINYINT(1) NOT NULL, INDEX IDX_3D8C939E8EBB8CC6 (pcd_partner), PRIMARY KEY(pcd_id)) DEFAULT CHARACTER SET utf8mb4 COLLATE `utf8mb4_unicode_ci` ENGINE = InnoDB');
        $this->addSql('ALTER TABLE promo_code ADD CONSTRAINT FK_3D8C939E8EBB8CC6 FOREIGN KEY (pcd_partner) REFERENCES user (id)');
        $this->addSql('ALTER TABLE exchange_symbol CHANGE exchange exchange VARCHAR(255) NOT NULL');
        $this->addSql('ALTER TABLE user ADD promo_code INT DEFAULT NULL, ADD partner_budget DOUBLE PRECISION NOT NULL');
        $this->addSql('ALTER TABLE user ADD CONSTRAINT FK_8D93D6493D8C939E FOREIGN KEY (promo_code) REFERENCES promo_code (pcd_id)');
        $this->addSql('CREATE INDEX IDX_8D93D6493D8C939E ON user (promo_code)');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE user DROP FOREIGN KEY FK_8D93D6493D8C939E');
        $this->addSql('ALTER TABLE promo_code DROP FOREIGN KEY FK_3D8C939E8EBB8CC6');
        $this->addSql('DROP TABLE promo_code');
        $this->addSql('ALTER TABLE exchange_symbol CHANGE exchange exchange CHAR(10) NOT NULL');
        $this->addSql('DROP INDEX IDX_8D93D6493D8C939E ON user');
        $this->addSql('ALTER TABLE user DROP promo_code, DROP partner_budget');
    }
}
