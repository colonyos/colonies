## HTTP RPC protocol
It is expected that the signature is set as a HTTP Header named **Signature**.

The Reply message can also contain an error message if the HTTP status code is not 200 OK.
```json
{
    "stats": "500",
    "name": "something when wrong here"
}
```

### Add Colony
Needs to be signed by a Server Owner Private Key.
#### Message

```json
{
    "rpc": {
        "method": "addcolony",
        "nonce": "5681b8c0e9f966df9b51e37e351449ad50a315baf20023ce0d24666dad59b991"
    },
    "colony": {
        "colonyid": "6d61afe7914c63f28a4c97645ce6ab264c3ad3a0e46ebd1f3788e83053934e18",
        "name": "test_colony_name"
    }
}
```

#### Reply 
```json
{
    "colonyid": "6d61afe7914c63f28a4c97645ce6ab264c3ad3a0e46ebd1f3788e83053934e18",
    "name": "test_colony_name"
}
```
