# IAM

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
make run
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

vault kv put -mount=secret app/global/gwebcredentials data=-
```

```txt
secret/data/app/global/gwebcredentials = {{
  "web": {
    "client_id": string,
    "project_id": string,
    "auth_uri": string,
    "token_uri": string,
    "auth_provider_x509_cert_url": string,
    "client_secret": string,
    "redirect_uris": []string
  }
}}
```
