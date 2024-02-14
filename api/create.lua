request = function()
    headers = {}
    headers["Content-Type"] = "application/json"
    body = '{ "name": "New wallet" }'
    return wrk.format("POST", "/wallets", headers, body)
end
