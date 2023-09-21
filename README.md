# iam

## Proto files

### Add a proto template

```bash
kratos proto add api/server/server.proto
```

### Generate the proto code

```bash
kratos proto client api/server/server.proto
```

### Generate the source code of service by proto file

```bash
kratos proto server api/server/server.proto -t internal/service

go generate ./...
```

## Generate other auxiliary files by Makefile

### Download and update dependencies

```bash
make init
```

### Generate API files (include: pb.go, http, grpc, validate, swagger) by proto file

```bash
make api
```

### Generate all files

```bash
make all
```

### Generate migrations

[Install Atlas](https://entgo.io/docs/versioned-migrations#generating-migrations)

```bash
make migrations
```

## Run

### Run debug

```bash
kratos run
```

### Build & Run

```bash
go build -o ./bin/ ./...
./bin/media -conf ./configs
```

## Run in Docker

```bash
make start
```

To stop docker:

```bash
make stop
```

## Vault

TODO: DB creds

To save JWT secret in Vault terminal (write command, ENTER, paste secret, CTRL+D):

```bash
export VAULT_TOKEN=myroot
vault kv put -mount=secret app/global/jwt data=-
```

## JWT token for your services

To get JWT token for any services, use API `/v1/auth/supercode` with code `sup3rcaL2033`
