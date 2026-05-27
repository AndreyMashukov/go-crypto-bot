<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

final class Version20240603050611 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        $users = $this->connection->executeQuery('SELECT id, roles FROM user')->fetchAllAssociative();
        foreach ($users as $user) {
            $roles = json_encode(unserialize($user['roles']));
            $this->addSql("UPDATE user SET roles = '{$roles}' WHERE id = '{$user['id']}'");
        }
        $this->addSql('ALTER TABLE user CHANGE roles roles JSON NOT NULL');
    }

    public function down(Schema $schema): void
    {
        $users = $this->connection->executeQuery('SELECT id, roles FROM user')->fetchAllAssociative();
        $this->addSql('ALTER TABLE user CHANGE roles roles LONGTEXT NOT NULL COMMENT \'(DC2Type:array)\'');
        foreach ($users as $user) {
            $roles = serialize(json_decode($user['roles'], true));
            $this->addSql("UPDATE user SET roles = '{$roles}' WHERE id = '{$user['id']}'");
        }
    }
}
