package icmp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// proto
const (
	ICMPv4 = 1
	ICMPv6 = 58
	//ipv4Proto = map[string]string{"ip": "ip4:icmp", "udp": "udp4"}
	//ipv6Proto = map[string]string{"ip": "ip6:ipv6-icmp", "udp": "udp6"}
)

// Lookup DNS
func Lookup(host string) (string, error) {
	addrs, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", errors.New("unknown host")
	}
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return addrs[rd.Intn(len(addrs))], nil
}

// EchoStat echo reply stat
type EchoStat struct {
	Seq  int
	TTL  int
	Cost time.Duration
}

// ICMP icmp config
type ICMP struct {
	opts Options
	ID   int
}

// New ping
func New(opts ...Option) *ICMP {
	options := newOptions(opts...)
	return &ICMP{
		opts: options,
		ID:   os.Getpid() & 0xFFFF,
	}
}

// Send without stdout
func (i *ICMP) Send(ctx context.Context, host string) (Statistics, error) {
	return i.ping(ctx, host, false)
}

// Ping ping
func (i *ICMP) Ping(ctx context.Context, host string) {
	_, _ = i.ping(ctx, host, true)
}

func (i *ICMP) ping(ctx context.Context, host string, print bool) (Statistics, error) {
	stats := Statistics{
		Host:   host,
		minRTT: math.MaxInt64,
		Loss:   100,
	}

	addr, err := Lookup(host)
	if err != nil {
		return stats, err
	}

	// if ip.To4() != nil {
	// 	proto = "v4"
	// 	network = "ip4:icmp"
	// } else {
	// 	proto = "v6"
	// 	network = "ip6:ipv6-icmp"
	// }
	if print {
		_, _ = fmt.Fprintf(os.Stdout, "PING %s (%s) %d(%d) bytes of data.\n", host, addr, i.opts.size, i.opts.size+28)
	}

Loop:
	for seq := 0; seq < i.opts.count; seq++ {
		if seq != 0 {
			time.Sleep(time.Second)
		}
		select {
		case <-ctx.Done():
			if print {
				_, _ = fmt.Fprintf(os.Stdout, "\n")
			}
			break Loop
		default:
			stat := &EchoStat{}
			stat, err = i.echo(ctx, addr, seq)
			stats.update(stat)
			if !print {
				continue
			}
			if err != nil {
				_, _ = fmt.Fprintf(os.Stdout, "%v\n", err)
			} else {
				_, _ = fmt.Fprintf(os.Stdout, "%d bytes from %s: icmp_seq=%d ttl=%d time=%s\n", i.opts.size, addr, stat.Seq, stat.TTL, stat.Cost)
			}
		}
	}
	if print {
		stats.Print(os.Stdout)
	}
	return stats, err
}

// Echo send one icmp packet
func (i *ICMP) Echo(ctx context.Context, addr string) (*EchoStat, error) {
	seq := int(time.Now().UnixNano()) % 65536
	return i.echo(ctx, addr, seq)
}

func (i *ICMP) echo(ctx context.Context, addr string, seq int) (*EchoStat, error) {
	if seq > 65535 {
		return nil, fmt.Errorf("Invalid ICMP Sequence number. Value must be 0<=N<2^16")
	}
	conn, err := net.Dial("udp:icmp", addr)
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
			Data: bytes.Repeat([]byte("x"), i.opts.size),
		},
	}
	payload, err := pkt.Marshal(nil)
	if err != nil {
		return nil, err
	}

	conn.SetDeadline(time.Now().Add(i.opts.timeout))

	now := time.Now()

	size, err := conn.Write(payload)
	if err != nil {
		return nil, fmt.Errorf("sendto: %v", err)
	}
	if size != len(payload) {
		return nil, errors.New("send ping data err")
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			ipbuf := make([]byte, 20+size) // ping recv back
			_, err := conn.Read(ipbuf)

			if err != nil {
				if opt, ok := err.(*net.OpError); ok && opt.Timeout() {
					return nil, fmt.Errorf("Request timeout for icmp_seq %d", seq)
				}
				return nil, fmt.Errorf("Read Fail: %w", err)
			}

			head, err := ipv4.ParseHeader(ipbuf)
			if err != nil {
				return nil, err
			}

			if head.Src.String() != addr {
				continue
			}

			msg, err := icmp.ParseMessage(ICMPv4, ipbuf[head.Len:])
			if err != nil {
				return nil, err
			}

			switch msg.Type {
			case ipv4.ICMPTypeEcho:
				continue
			case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
				echo, ok := msg.Body.(*icmp.Echo)
				if !ok {
					return nil, errors.New("ping recv err reply data")
				}
				if echo.ID != msg.Body.(*icmp.Echo).ID || seq != echo.Seq {
					continue
				}
				return &EchoStat{echo.Seq, head.TTL, time.Since(now)}, nil

			case ipv4.ICMPTypeDestinationUnreachable:
				return nil, fmt.Errorf("From %s icmp_seq=%d Destination Unreachable", addr, seq)
			case ipv4.ICMPTypeTimeExceeded:
				_, ok := msg.Body.(*icmp.TimeExceeded)
				if !ok {
					return nil, errors.New("ping recv err reply data")

				}
				return nil, errors.New("TimeExceeded")
				// if len(rply.Data) > 24 {
				// 	if uint16(seq) == binary.BigEndian.Uint16(rply.Data[24:26]) {
				// 		return &EchoStat{seq, head.TTL, time.Since(now)}
				// 	}
				// }
			default:
				return nil, fmt.Errorf("Not ICMPTypeEchoReply seq=%d, %#v, %#v", seq, msg, msg.Body)
			}
		}
	}
}
