// Package unqid provides a distributed unique id generator
//
// A unqid ID is composed of
//     40 bits for time in ms since 2021-01-01
//     16 bits for a machine id generated from lower 16 bits of the first private network address
//      7 bits for a sequence number
//
// unqid will panic during initialization if:
// 1. No private ip address could be found to generate machine id bits
// 2. If current time is past the supported max time (2055-11-04 19:53:47.775 +0000 UTC)
package unqid

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

const (
	seqBits       = 7                        // sequence bits
	machBits      = 16                       // machine id bits
	machBitsShift = seqBits                  // number of bits to shift for machine ID
	tsBits        = 41                       // timestamp bits
	tsBitsShift   = machBits + machBitsShift // timestamp bits shift
	maxSeq        = (1 << seqBits) - 1       // 2^seqBits - 1
	maxTS         = (1 << tsBits) - 1        // 2^tsBits - 1
)

var (
	epoch, _ = time.Parse("2006-01-02 15:04:05", "2021-01-01 00:00:00")
	machID   = int64(0)
	seq      = int64(0)
	elapsed  = int64(0)

	since          = time.Since
	interfaceAddrs = net.InterfaceAddrs
)

func init() {
	initialize()
}

func initialize() {
	mid, err := machineID()
	if err != nil {
		panic(err)
	}
	if mid == 0 {
		panic("invalid machine id: 0")
	}
	machID = mid << machBitsShift

	e := since(epoch).Milliseconds()
	if e > maxTS {
		panic("max time exceeded")
	}
	atomic.StoreInt64(&elapsed, since(epoch).Milliseconds())
}

func machineID() (int64, error) {
	addrs, err := interfaceAddrs()
	if err != nil {
		return 0, err
	}
	for _, a := range addrs {
		ipn, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipn.IP
		ipBytes := ip.To4()
		if ipBytes != nil && !ip.IsLoopback() && ip.IsPrivate() {
			return int64(uint16(ipBytes[2])<<8 + uint16(ipBytes[3])), nil
		}
	}
	return 0, errors.New("no private ip address")
}

// Next returns the next unique ID.
//
// Calling Next will result in a panic after the supported max time (2055-11-04 19:53:47.775 +0000 UTC).
// If the system clock is rolledback to a time before the last generated ID (e.g. using NTP), Next will wait until clock exceed the last generated timestamp before returning a new ID.
// If number of IDs generated in the current millisecond exceeds max supported (128), Next will wait till the next sequence reset (done every ms).
func Next() int64 {
	t := since(epoch).Milliseconds()

	// check end of life
	if t > maxTS {
		panic(fmt.Sprintf("max time exceeded: %v, maxtime: %v", t, maxTS))
	}

	// check clock rollbacks
	e := atomic.LoadInt64(&elapsed)
	for e > t {
		t = since(epoch).Milliseconds()
		e = atomic.LoadInt64(&elapsed)
	}

	// check sequence overruns
	s := atomic.AddInt64(&seq, 1)
	for s > maxSeq {
		for t == e {
			t = since(epoch).Milliseconds()
		}
		atomic.StoreInt64(&seq, 0)
		s = 0
	}
	atomic.StoreInt64(&elapsed, t)

	return (t << tsBitsShift) | machID | s
}
