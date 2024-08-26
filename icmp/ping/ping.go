package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/corex-io/libnet/icmp"
)

func main() {
	count := flag.Int("c", 4, "count")
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

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)

	ping := icmp.New(
		icmp.Count(*count),
		icmp.Size(*pktsize),
		icmp.Timeout(time.Duration(*timeout)*time.Second),
		icmp.Log(os.Stdout),
	)

	stats, err := ping.Ping(context.Background(), hosts[0])
	if err != nil {
		fmt.Println("ping", hosts[0], "fail:", err)
	}
	fmt.Println(stats.Print())

}
