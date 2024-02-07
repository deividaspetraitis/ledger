#!/bin/sh

curl --json '{ "transaction": "withdraw", "wallet_id": "5a486373-14d8-4643-ad9b-2cadc77f7a98", "amount": 150 }' http://localhost/transactions -v
