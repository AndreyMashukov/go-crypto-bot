lint:
	/Users/amashukov/go/bin/golint .
init-db-dev:
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_1.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_2.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_3.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_4.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_5.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_6.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_7.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_8.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_9.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_10.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_11.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_12.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_13.sql
