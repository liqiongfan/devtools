package icmp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

var (
	MaxPg       = 2000
	originBytes []byte
)

type icmp struct {
	Type        uint8
	Code        uint8
	CheckSum    uint16
	Identifier  uint16
	SequenceNum uint16
}

func init() {
	originBytes = make([]byte, MaxPg)
}

// checkSum
func checkSum(data []byte) (rt uint16) {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index]) << 8
	}

	rt = uint16(sum) + uint16(sum>>16)
	return ^rt
}

type pingOpt struct {
	domain    string
	ps, count int
	echo      bool
}
type PingOption func(*pingOpt)

func WithDomain(domain string) PingOption {
	return func(opt *pingOpt) {
		opt.domain = domain
	}
}

const (
	MacPs     = 28
	WindowsPs = 4
	LinuxPs   = 56
)

// WithPs on Mac ps size need be 28
// WithPs on Windows ps size should be 4
// WithPs on Linux ps size should be 56
func WithPs(ps int) PingOption {
	return func(opt *pingOpt) {
		opt.ps = ps
	}
}

func WithCount(count int) PingOption {
	return func(opt *pingOpt) {
		opt.count = count
	}
}

func WithEcho(echo bool) PingOption {
	return func(opt *pingOpt) {
		opt.echo = echo
	}
}

// Ping Get the ping result and return DropBack rate
// result ends with 0.000000
func Ping(opts ...PingOption) float64 {
	opt := &pingOpt{
		domain: "127.0.0.1",
		ps:     32,
		count:  4,
		echo:   true,
	}

	for _, o := range opts {
		o(opt)
	}

	domain := opt.domain
	ps := opt.ps
	count := opt.count
	echo := opt.echo

	var (
		message                            icmp
		localAddr                          = net.IPAddr{IP: net.ParseIP("0.0.0.0")}
		remoteAddr, _                      = net.ResolveIPAddr("ip", domain)
		maxLatency, minLatency, avgLatency float64
		buffer                             bytes.Buffer
	)

	c, err := net.DialIP("ip4:icmp", &localAddr, remoteAddr)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = c.Close()
	}()

	message = icmp{8, 0, 0, 0, 0}

	_ = binary.Write(&buffer, binary.BigEndian, message)
	_ = binary.Write(&buffer, binary.BigEndian, originBytes[0:ps])
	b := buffer.Bytes()
	binary.BigEndian.PutUint16(b[2:], checkSum(b))

	if echo {
		fmt.Printf("\n正在 Ping %s 具有 %d(%d) 字节的数据:\n", remoteAddr.String(), ps, ps+28)
	}

	responseBuff := make([]byte, 1024)
	var responseList []float64

	dropPack := 0.0
	maxLatency = 3000.0
	minLatency = 0.0
	avgLatency = 0.0

	for i := count; i > 0; i-- {
		if _, err := c.Write(buffer.Bytes()); err != nil {
			dropPack++
			continue
		}

		startTime := time.Now()
		_ = c.SetReadDeadline(time.Now().Add(time.Second))
		l, err := c.Read(responseBuff)
		if err != nil {
			dropPack++
			continue
		}

		endTime := time.Now()
		duration := float64(endTime.Sub(startTime).Nanoseconds()) / 1e6
		responseList = append(responseList, duration)
		if duration < maxLatency {
			maxLatency = duration
		}

		if duration > minLatency {
			minLatency = duration
		}
		if echo {
			fmt.Printf("来自 %s 的回复: 大小 = %d byte 时间 = %.3fms\n", remoteAddr.String(), l, duration)
		}
	}

	if echo {
		fmt.Printf("丢包率: %.2f%%\n", dropPack/float64(count)*100)
	}
	if len(responseList) == 0 {
		avgLatency = 3000.0
	} else {
		sum := 0.0
		for _, n := range responseList {
			sum += n
		}
		avgLatency = sum / float64(len(responseList))
	}

	if echo {
		fmt.Printf("rtt 最短 = %.3fms 平均 = %.3fms 最长 = %.3fms\n", minLatency, avgLatency, maxLatency)
	}

	return dropPack / float64(count) * 100
}
