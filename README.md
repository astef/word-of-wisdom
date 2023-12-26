# Word of Wisdom

Demo TCP server protected from DDoS with PoW challenge-response protocol

## PoW algorithm

In the terminology of the [Hashcash - A Denial of Service Counter-Measure](http://www.hashcash.org/hashcash.pdf) paper, we are using **hashcash-cookie** algorithm.

**SHA256** is used as an _interactive cost function_.

**HMACSHA256** is used for challenge verification on the server.

Important distinctive features of this implementation are:

- Protocol is designed to be stateless, just like HTTP request-response, because the number TCP connections is critical for resisting DDoS
- Challenge includes expiration time, so an attacker's ability to accumulate solutions is limited, as is the risk of DoS. Moreover, this nice feature let's the server to dynamically change algorithm parameters and loose support of old challenges by just waiting for them to expire.
- Client IP address is added to the challenge signature by the server, so it is impossible to change it between the challenge-response calls, so an attacker won't be able to accumulate solutions from different locations and spend them all in one location with the purpose of DoS
- It has zero dependencies, and quite low amount of source code: 10 files, <600 non-blank non-comment lines

## Running

```
docker compose up --force-recreate --build server --build client
```

## Known problems and limitations

- Server secret is re-generated on every start, so challenges should not survive server restart, or be used with different servers
- It fails with Go `1.21.0`, because of the [bug](https://github.com/golang/go/issues/62117), please use Go `1.21.5` or higher
- Message format is determined by `encoding/gob`, which is not interoperable format. This is done intentionally to simplify the implementation. It is easily replaceable with a more advanced formatters.
