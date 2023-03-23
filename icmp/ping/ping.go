package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/corex-io/libnet/icmp"
)

func main() {
	count := flag.Int("c", 0, "count")
	pktsize := flag.Int("s", 56, "packetsize")
	timeout := flag.Int("t", 5, "timeout")
	flag.Parse()
	args := flag.Args()
	for idx, pos := range args {
		if strings.HasPrefix(pos, "-") {
			args[idx] = ""
			if idx+1 != len(args) {
				args[idx+1] = ""
			}
		}
	}

	var hosts []string

	for _, arg := range args {
		if arg != "" {
			hosts = append(hosts, arg)
		}
	}

	if len(hosts) == 0 {
		fmt.Println("parse hosts fail", hosts)
		return
	}

	var wg sync.WaitGroup
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)

	ctx, cancel := context.WithCancel(context.Background())
	ping := icmp.New(icmp.Count(*count), icmp.Size(*pktsize), icmp.Timeout(time.Duration(*timeout)*time.Second))

	ch2 := make(chan struct{}, len(hosts))

	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			ping.Ping(ctx, host)

		}(host)
	}

	select {
	case <-ch:
		cancel()
	case <-ch2:
	}
	wg.Wait()

}
