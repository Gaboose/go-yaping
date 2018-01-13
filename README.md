# go-yaping

Yet another ICMP Ping library for Go. Inspired by [github.com/sparrc/go-ping](https://github.com/sparrc/go-ping)

Yaping follows two principles:

1. Use `p.Ping()` to issue a ping and to return the round-trip-time (not via callback or channel).
2. Don't create any goroutines inside the library (i.e. opaque to the user) to avoid complicated teardown.

Install with `go get github.com/Gaboose/go-yaping`

## Simple example

```go
conn, _ := icmp.ListenPacket("ip4:icmp", "")
ipaddr, _ := net.ResolveIPAddr("ip4", "github.com")

p := yaping.Pinger{
    Conn: conn,
    Addr: ipaddr,
    ID:   uint16(rand.Intn(65536)),
}

go yaping.Listener{conn}.Listen(func(addr net.Addr, pkt *icmp.Echo) {
    if uint16(pkt.ID) == p.ID {
        p.Accept(pkt)
    }
})

rtt, seq, err := p.Ping()
```

## Better example

Command-line tool [cmd/yaping/main.go](https://github.com/Gaboose/go-yaping/blob/master/cmd/yaping/main.go)

```bash
$ yaping
Usage:
	yaping host [host [host [...]]]
Example:
	yaping google.com wikipedia.org

$ sudo yaping google.com wikipedia.org
from 172.217.21.14: seq=0 rtt=29.054002ms
from 91.198.174.192: seq=0 rtt=38.503945ms
from 172.217.21.14: seq=1 rtt=28.807146ms
from 91.198.174.192: seq=1 rtt=37.774642ms
...
```

Install with `go get github.com/Gaboose/go-yaping/cmd/yaping`

## Note on Privileges

Unprivileged ping messages from userspace are problematic. This library only works if the appliation is run from a superuser account.

See [github.com/sparrc/go-ping#note-on-linux-support](https://github.com/sparrc/go-ping#note-on-linux-support) for more details.