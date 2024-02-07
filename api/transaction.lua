types = { "deposit", "withdraw", "unsupported" }
request = function()
    headers = {}
    headers["Content-Type"] = "application/json"
    body = '{ "transaction": "' .. types[math.random(#types)] .. '", "wallet_id": "c8082cb1-9440-41f1-9358-9ee94f2ba838", "amount": ' .. math.random(1,1000) .. ' }'
    return wrk.format("POST", "/transactions", headers, body)
end
