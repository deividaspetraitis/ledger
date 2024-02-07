# Description

The goal is to build a simple high throughput ledger service supporting various transaction operations.

## Documentation

One of possible ways to read the project and available documentation is by using `pkgsite`.

```bash
go install golang.org/x/pkgsite/cmd/pkgsite@latest
```

Considering you're in the project root:

```bash
pkgsite -open .
```

## Functional API description

Service expose following API HTTP endpoints 

### POST /wallets
Creates a new wallet.

Send a request to the running service instance ( presuming service is running on port `80` ):

```bash
curl --json '{ "name": "Family Fund" }' http://localhost/wallets -v
```
#### HTTP 200 

Successful request response example:

```json
{"id":"a18c247b-8c28-468f-97a8-0bf33a48b922","name":"Family Fund","balance":0}
```

#### HTTP 500 

All other non-successful requests will return `HTTP 500` with empty body.

### POST /transactions
Add or remove funds from the wallet. Amount is specified in cents. Supported transactions are following:

* deposit
* withdraw

Send a request to the running service instance ( presuming service is running on port `80` ):

```bash
curl --json '{ "transaction": "deposit", "wallet_id": "a18c247b-8c28-468f-97a8-0bf33a48b922", "amount": 150 }' http://localhost/transactions -v
```

#### HTTP 200 

Successful request response example:

```json
{"transaction":"deposit","wallet_id":"a18c247b-8c28-468f-97a8-0bf33a48b922","amount":150}
```

#### HTTP 500 

All other non-successful requests will return `HTTP 500` with empty body.

### GET /wallet/{wallet_id}
Query the current state of the wallet.

Send a request to the running service instance ( presuming service is running on port `80` ):

```bash
curl http://localhost/wallets/cbd1c9a2-95fc-4a6e-a5fe-f7da7e4362b3 -v
```

Successful request response example:

```json
{"id":"a18c247b-8c28-468f-97a8-0bf33a48b922","name":"Family Fund","balance":150}
```

#### HTTP 404 

If given wallet is not found request will result in `HTTP 404` with empty body.

#### HTTP 500 

All other non-successful requests will return `HTTP 500` with empty body.

### Responses

All other non successful responses at this moment are returned as HTTP 500 for simplicity.

## Requirements

* We need a way to create a wallet
* We need a way to add funds to a wallet
* We need a way to remove funds from the wallet
* We need a way to query the current state of the wallet
* The balance of the wallet can not be negative
* It's of utmost importance that user can not spend same funds twice
* This service is part of critical and time sensitive workflows, so performance is important
* The service should be able to shut down gracefully

## Technical overview

### Implementation highlights

### Possible improvements:

This application as any other can be improved in many different ways and is far from perfect. Several good improvement ideas might be:

* Proper HTTP status codes and responses for non-successful requests, e.g. wallet not found, validation errors
* Increase tests coverage
* Logs redirect in test mode
* Resolve TODOs
* Use more convenient way to document APIs ( Postman decided to stop shipping past releases, and now requires registered accounts )
* ...
