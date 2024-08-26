package icmp

import (
	"bytes"
	"fmt"
	"math"
	"time"
)

// Statistics ..
type Statistics struct {
	Host        string
	Transmitted int
	Received    int
	Loss        float64
	Time        time.Duration
	TimeExp2    int64
	minRTT      time.Duration
	avgRTT      time.Duration
	maxRTT      time.Duration
	mdevRTT     time.Duration
}

func (stats *Statistics) update(stat *EchoStat) {
	stats.Transmitted++
	if stat != nil {
		stats.Received++
		stats.Time += stat.Cost

		stats.TimeExp2 += stat.Cost.Nanoseconds() * stat.Cost.Nanoseconds()
		if stats.minRTT > stat.Cost {
			stats.minRTT = stat.Cost
		}
		stats.avgRTT = time.Duration(stats.Time.Nanoseconds()/int64(stats.Received)) * time.Nanosecond
		if stats.maxRTT < stat.Cost {
			stats.maxRTT = stat.Cost
		}
		stats.mdevRTT = time.Duration(math.Sqrt(float64(stats.TimeExp2)/float64(stats.Received)-float64(stats.avgRTT.Nanoseconds()*stats.avgRTT.Nanoseconds()))) * time.Nanosecond
	}
	stats.Loss = float64((stats.Transmitted-stats.Received)*100) / float64(stats.Transmitted)
}

func (stats *Statistics) String() string {
	return fmt.Sprintf("PING %s: %d packets transmitted, %d received, %.2f%% packet loss",
		stats.Host, stats.Transmitted, stats.Received, stats.Loss)
}

// Print print
func (stats *Statistics) Print() string {
	buf := &bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("--- %s ping statistics ---\n", stats.Host))
	buf.WriteString(fmt.Sprintf("%d packets transmitted, %d received, %.2f%% packet loss, time %s\n",
		stats.Transmitted, stats.Received, stats.Loss, stats.Time))
	if stats.Received == stats.Transmitted {
		buf.WriteString(fmt.Sprintf("rtt min/avg/max/mdev = %.3f/%.3f/%.3f/%.3f ms\n",
			stats.minRTT.Seconds()*1000, stats.avgRTT.Seconds()*1000, stats.maxRTT.Seconds()*1000, stats.mdevRTT.Seconds()*1000))
	}
	return buf.String()
}
