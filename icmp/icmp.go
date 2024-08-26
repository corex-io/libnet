package icmp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// proto
const (
	ICMPv4 = 1
	ICMPv6 = 58
	//ipv4Proto = map[string]string{"ip": "ip4:icmp", "udp": "udp4"}
	//ipv6Proto = map[string]string{"ip": "ip6:ipv6-icmp", "udp": "udp6"}
)

// EchoStat echo reply stat
type EchoStat struct {
	Seq  int
	TTL  int
	Cost time.Duration
}

// ICMP icmp config
type ICMP struct {
	opts []Option
	ID   int
}

// New ping
func New(opts ...Option) *ICMP {
	return &ICMP{
		opts: opts,
		ID:   os.Getpid() & 0xFFFF,
	}
}

// Ping without stdout
func (i *ICMP) Ping(ctx context.Context, host string, opts ...Option) (*Statistics, error) {
	options := newOptions(i.opts, opts...)
	stats := &Statistics{Host: host, minRTT: -1, Loss: -1}

	addr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		fmt.Println(err.Error())
		//ping: cannot resolve qq.com: Unknown host
		return stats, err
	}

	_, _ = fmt.Fprintf(options.Log, "PING %s (%s) %d(%d) bytes of data.\n", host, addr, options.size, options.size+28)

	for seq := 0; seq < options.count; seq++ {
		select {
		case <-ctx.Done():
			_, _ = fmt.Fprintf(options.Log, "\n")
			return stats, ctx.Err()
		default:
			stat, err := i.echo(ctx, addr, seq, opts...)
			if err != nil {
				return stats, err
			}
			stats.update(stat)

			_, _ = fmt.Fprintf(options.Log, "%d bytes from %s: icmp_seq=%d ttl=%d time=%.3fms\n", options.size, addr, stat.Seq, stat.TTL, stat.Cost.Seconds()*1000)

			time.Sleep(1 * time.Second)

		}
	}

	return stats, err
}

func (i *ICMP) echo(ctx context.Context, addr *net.IPAddr, seq int, opts ...Option) (*EchoStat, error) {
	options := newOptions(i.opts, opts...)

	seq = seq % 65536

	conn, err := net.DialIP("ip4:icmp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("dial fail: %v", err)
	}
	defer conn.Close()

	pkt := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   i.ID,
			Seq:  seq,
			Data: bytes.Repeat([]byte("x"), options.size),
		},
	}
	content, err := pkt.Marshal(nil)
	if err != nil {
		return nil, err
	}

	if err := conn.SetDeadline(time.Now().Add(options.timeout)); err != nil {
		return nil, err
	}

	now := time.Now()

	size, err := conn.Write(content)
	if err != nil || size != len(content) {
		return nil, fmt.Errorf("sendto: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			ipbuf := make([]byte, 20+size) // ping recv back
			_, err := conn.Read(ipbuf)

			if err != nil {
				var opt *net.OpError
				if errors.As(err, &opt) && opt.Timeout() {
					return nil, fmt.Errorf("request timeout for icmp_seq %d", seq)
				}
				return nil, fmt.Errorf("read Fail: %w", err)
			}

			head, err := ipv4.ParseHeader(ipbuf)
			if err != nil {
				return nil, err
			}

			if head.Src.String() != addr.String() {
				continue
			}

			msg, err := icmp.ParseMessage(ICMPv4, ipbuf[head.Len:])
			if err != nil {
				return nil, err
			}

			switch msg.Type {
			case ipv4.ICMPTypeEcho:
				continue
			case ipv4.ICMPTypeEchoReply:
				echo, ok := msg.Body.(*icmp.Echo)
				if !ok {
					return nil, errors.New("ping recv err reply data")
				}
				if echo.ID != msg.Body.(*icmp.Echo).ID || seq != echo.Seq {
					continue
				}
				return &EchoStat{echo.Seq, head.TTL, time.Since(now)}, nil

			case ipv4.ICMPTypeDestinationUnreachable:
				return nil, fmt.Errorf("from %s icmp_seq=%d Destination Unreachable", addr, seq)
			case ipv4.ICMPTypeTimeExceeded:
				_, ok := msg.Body.(*icmp.TimeExceeded)
				if !ok {
					return nil, errors.New("ping recv err reply data")

				}
				return nil, errors.New("TimeExceeded")
			default:
				return nil, fmt.Errorf("not ICMPTypeEchoReply seq=%d, %#v, %#v", seq, msg, msg.Body)
			}
		}
	}
}
