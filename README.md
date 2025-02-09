# basic-golang-api
Basic API with Golang

## Set PATH

```bash
export PATH=$PATH:/usr/local/go/bin
```

## Running app

```bash
go run main.go
```

## Build app

```bash
go build -o ./bin
```

## Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin", "password":"password123"}'
```

## Insert book

```bash
curl -X POST http://localhost:8080/books \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id":"4", "title":"The Hobbit", "author":"J.R.R. Tolkien"}'
```

## Build Golang app

```bash
go build -o bin
```

## Run on Docker

```bash
docker-compose up -d
```