.PHONY: fmt vet lint ailint build test check init-db-dev

# Go targets — module root is this directory.

fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	golangci-lint run --config=../.golangci.yml --timeout=3m

ailint:
	go-lint ./...

build:
	go build ./...

test:
	go test -count=1 -race -timeout=120s ./...

check: fmt vet build lint ailint test

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
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_14.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_15.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_16.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_17.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_18.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_19.sql
	mysql -u root -pgo_crypto_bot -h 127.0.0.1 -P 3367 -D go_crypto_bot < migrations/migration_20.sql
