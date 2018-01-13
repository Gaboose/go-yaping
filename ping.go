package yaping

import (
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const DefaultTimeout = 10 * time.Second

type Error struct {
	addr net.Addr
	seq  uint16
}

func (e Error) Error() string {
	return fmt.Sprintf("Timeout from %s: seq=%d", e.addr, e.seq)
}

func (e Error) Timeout() bool {
	return true
}

type Listener struct {
	Conn net.PacketConn
}

// https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
const protocolICMP = 1

func (pl Listener) Listen(fn func(net.Addr, *icmp.Echo)) {
	bytes := make([]byte, 512)
	for {
		_, addr, err := pl.Conn.ReadFrom(bytes)
		if err != nil {
			panic(err)
		}
		var m *icmp.Message
		if m, err = icmp.ParseMessage(protocolICMP, bytes); err != nil {
			panic(err)
		}

		pkt, ok := m.Body.(*icmp.Echo)
		if !ok {
			// ignore everything that's not an echo reply
			continue
		}
		fn(addr, pkt)
	}
}

type Pinger struct {
	Conn    net.PacketConn
	Addr    net.Addr
	ID      uint16
	Timeout time.Duration

	seq uint16
	ch  map[uint16]chan *icmp.Echo
	mu  sync.Mutex
}

func (p *Pinger) Ping() (rtt time.Duration, seq uint16, err error) {

	timeout := p.Timeout
	if p.Timeout == 0 {
		timeout = DefaultTimeout
	}
	ch := make(chan *icmp.Echo, 1)

	p.mu.Lock()
	seq = p.seq
	if p.ch == nil {
		p.ch = map[uint16]chan *icmp.Echo{}
	}
	p.ch[p.seq] = ch
	p.seq++
	p.mu.Unlock()

	start := time.Now()

	err = sendIcmpEcho(p.Conn, p.Addr, &icmp.Echo{
		ID:  int(p.ID),
		Seq: int(seq),
	})
	if err != nil {
		return
	}

	select {
	case <-ch:
		rtt = time.Now().Sub(start)
	case <-time.After(timeout):
		err = Error{addr: p.Addr, seq: seq}
	}

	p.mu.Lock()
	delete(p.ch, seq)
	p.mu.Unlock()

	return
}

func (p *Pinger) Accept(pkt *icmp.Echo) {
	p.mu.Lock()
	ch, ok := p.ch[uint16(pkt.Seq)]
	p.mu.Unlock()
	if ok {
		ch <- pkt
	}
}

func sendIcmpEcho(conn net.PacketConn, ipaddr net.Addr, body *icmp.Echo) error {
	bytes, err := (&icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: body,
	}).Marshal(nil)
	if err != nil {
		return err
	}
	_, err = conn.WriteTo(bytes, ipaddr)
	return err
}
