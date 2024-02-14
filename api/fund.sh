#!/bin/sh

curl --json '{ "transaction": "deposit", "wallet_id": "5a486373-14d8-4643-ad9b-2cadc77f7a98", "amount": 150 }' http://localhost:8000/transactions -v
