<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20240117083926 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TABLE server (srv_id INT AUTO_INCREMENT NOT NULL, srv_ip VARCHAR(255) NOT NULL, srv_slots INT NOT NULL, srv_cpu INT NOT NULL, srv_ram INT NOT NULL, srv_master TINYINT(1) NOT NULL, PRIMARY KEY(srv_id)) DEFAULT CHARACTER SET utf8mb4 COLLATE `utf8mb4_unicode_ci` ENGINE = InnoDB');
        $this->addSql('INSERT INTO server (srv_ip, srv_slots, srv_cpu, srv_ram, srv_master) SELECT ctb_ip_address as ip, 3 as slots, 1 as `cpu`, 1 as ram, 1 as `master` FROM crypto_bot GROUP BY ctb_ip_address');
        $this->addSql('ALTER TABLE crypto_bot ADD ctb_server INT DEFAULT NULL, ADD ctb_master TINYINT(1) NOT NULL');
        $this->addSql('UPDATE crypto_bot b INNER JOIN server s ON b.ctb_ip_address = s.srv_ip SET b.ctb_server = s.srv_id');
        $this->addSql('ALTER TABLE crypto_bot DROP ctb_ip_address');
        $this->addSql('CREATE TABLE trade (trd_id INT AUTO_INCREMENT NOT NULL, trd_user INT NOT NULL, trd_bot INT NOT NULL, trd_profit DOUBLE PRECISION NOT NULL, trd_symbol VARCHAR(10) NOT NULL, trd_buy_qty DOUBLE PRECISION NOT NULL, trd_sell_qty DOUBLE PRECISION NOT NULL, trd_buy_price DOUBLE PRECISION NOT NULL, trd_sell_price DOUBLE PRECISION NOT NULL, trd_percent DOUBLE PRECISION NOT NULL, trd_buy_date DATETIME NOT NULL COMMENT \'(DC2Type:datetime_immutable)\', trd_sell_date DATETIME NOT NULL COMMENT \'(DC2Type:datetime_immutable)\', trd_order_id INT NOT NULL, INDEX IDX_7E1A4366B33E9D02 (trd_user), INDEX IDX_7E1A4366C22394F6 (trd_bot), PRIMARY KEY(trd_id)) DEFAULT CHARACTER SET utf8mb4 COLLATE `utf8mb4_unicode_ci` ENGINE = InnoDB');
        $this->addSql('CREATE TABLE transaction (trx_id INT AUTO_INCREMENT NOT NULL, trx_user INT NOT NULL, trx_amount DOUBLE PRECISION NOT NULL, trx_created_at DATETIME NOT NULL COMMENT \'(DC2Type:datetime_immutable)\', INDEX IDX_723705D1C72A5FE2 (trx_user), PRIMARY KEY(trx_id)) DEFAULT CHARACTER SET utf8mb4 COLLATE `utf8mb4_unicode_ci` ENGINE = InnoDB');
        $this->addSql('ALTER TABLE trade ADD CONSTRAINT FK_7E1A4366B33E9D02 FOREIGN KEY (trd_user) REFERENCES user (id)');
        $this->addSql('ALTER TABLE trade ADD CONSTRAINT FK_7E1A4366C22394F6 FOREIGN KEY (trd_bot) REFERENCES crypto_bot (ctb_id)');
        $this->addSql('ALTER TABLE transaction ADD CONSTRAINT FK_723705D1C72A5FE2 FOREIGN KEY (trx_user) REFERENCES user (id)');
        $this->addSql('ALTER TABLE crypto_bot ADD CONSTRAINT FK_BE990B7AD8E90F11 FOREIGN KEY (ctb_server) REFERENCES server (srv_id)');
        $this->addSql('CREATE INDEX IDX_BE990B7AD8E90F11 ON crypto_bot (ctb_server)');
        $this->addSql('ALTER TABLE user ADD budget DOUBLE PRECISION NOT NULL');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('ALTER TABLE crypto_bot DROP FOREIGN KEY FK_BE990B7AD8E90F11');
        $this->addSql('ALTER TABLE trade DROP FOREIGN KEY FK_7E1A4366B33E9D02');
        $this->addSql('ALTER TABLE trade DROP FOREIGN KEY FK_7E1A4366C22394F6');
        $this->addSql('ALTER TABLE transaction DROP FOREIGN KEY FK_723705D1C72A5FE2');
        $this->addSql('DROP TABLE server');
        $this->addSql('DROP TABLE trade');
        $this->addSql('DROP TABLE transaction');
        $this->addSql('DROP INDEX IDX_BE990B7AD8E90F11 ON crypto_bot');
        $this->addSql('ALTER TABLE crypto_bot ADD ctb_ip_address VARCHAR(255) DEFAULT NULL, DROP ctb_server, DROP ctb_master');
        $this->addSql('ALTER TABLE user DROP budget');
    }
}
