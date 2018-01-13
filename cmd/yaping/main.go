package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/Gaboose/go-yaping"

	"golang.org/x/net/icmp"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var usage = `
Usage:
	yaping host [host [host [...]]]
Example:
	yaping google.com wikipedia.org
`

func main() {

	hosts := os.Args[1:]
	if len(hosts) == 0 {
		fmt.Print(usage)
		return
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	pingers := make(map[uint16]*yaping.Pinger)

	for _, host := range hosts {
		ipaddr, err := net.ResolveIPAddr("ip4", host)
		if err != nil {
			log.Fatal(err)
		}

		p := yaping.Pinger{
			Conn: conn,
			Addr: ipaddr,
			ID:   uint16(rand.Intn(65536)),
		}

		pingers[p.ID] = &p
	}

	go yaping.Listener{conn}.Listen(func(addr net.Addr, pkt *icmp.Echo) {
		if p, ok := pingers[uint16(pkt.ID)]; ok {
			p.Accept(pkt)
		}
	})

	for {
		for _, p := range pingers {
			go func(p *yaping.Pinger) {
				dur, seq, err := p.Ping()
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("from %s: seq=%d rtt=%s\n", p.Addr, seq, dur)
			}(p)
		}
		time.Sleep(time.Second)
	}

}
