package unqid

import (
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestInitPanics(t *testing.T) {
	origEpoch := epoch
	defer func() {
		interfaceAddrs = net.InterfaceAddrs
		epoch = origEpoch
		initialize()
	}()

	t.Run("machineID error", func(t *testing.T) {
		defer func() { _ = recover() }()
		interfaceAddrs = func() ([]net.Addr, error) {
			return nil, errors.New("machine id error")
		}
		initialize()
		t.Errorf("should have panicked")
	})

	t.Run("machineID zero", func(t *testing.T) {
		defer func() { _ = recover() }()
		interfaceAddrs = func() ([]net.Addr, error) {
			_, ipn, err := net.ParseCIDR("192.168.0.0/24")
			return []net.Addr{ipn}, err
		}
		initialize()
		t.Errorf("should have panicked")
	})

	t.Run("max time exceeded", func(t *testing.T) {
		defer func() { _ = recover() }()
		interfaceAddrs = net.InterfaceAddrs
		epoch = time.Now().Add(time.Duration(-1*(1<<tsBits)) * time.Millisecond)
		initialize()
		t.Errorf("should have panicked")
	})

}

func Test_machineID(t *testing.T) {
	defer func() {
		interfaceAddrs = net.InterfaceAddrs
	}()

	tests := []struct {
		name    string
		ipnet   string
		want    int64
		wantErr bool
	}{
		{"valid1", "192.168.1.210/32", (1<<8 + 210), false},
		{"loopback", "127.0.0.1/8", 0, true},
		{"no-private", "77.73.33.222/32", 0, true},
		{"valid2", "10.10.10.201/24", (10 << 8), false},
		{"invalid", "22/22", 0, true},
	}
	lb := &net.IPAddr{IP: []byte{127, 0, 0, 1}, Zone: ""}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ipnet, err := net.ParseCIDR(tt.ipnet)
			interfaceAddrs = func() ([]net.Addr, error) {
				if err != nil {
					return nil, err
				}
				ips := []net.Addr{lb, ipnet}
				return ips, err
			}
			got, err := machineID()
			if (err != nil) != tt.wantErr {
				t.Errorf("machineID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("machineID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNext(t *testing.T) {
	machID = 336 << machBitsShift
	atomic.StoreInt64(&elapsed, 0)
	defer func() {
		initialize()
	}()
	tests := []struct {
		name        string
		elapsed     int64
		sequence    int64
		want        int64
		expectPanic bool
	}{
		{"maxTime", maxTS + 1, 0, 0, true},
		{"maxSequence", 100, maxSeq + 1, (101 << tsBitsShift) | machID, false},
		{"valid1", 44725615047, 0, (44725615047 << tsBitsShift) | machID | 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := time.Now()
			since = func(time.Time) time.Duration {
				return time.Duration(tt.elapsed+time.Since(st).Milliseconds()) * time.Millisecond
			}
			atomic.StoreInt64(&elapsed, tt.elapsed)
			atomic.StoreInt64(&seq, tt.sequence)
			if tt.expectPanic {
				defer func() {
					_ = recover()
				}()
			}
			if got := Next(); got != tt.want {
				t.Errorf("Next() = %v, want %v", got, tt.want)
				if tt.expectPanic {
					t.Errorf("Next() should have panicked")
				}
			}
		})
	}
}

func TestNextRollback(t *testing.T) {
	next1 := Next() >> (seqBits + machBits)
	next2 := Next() >> (seqBits + machBits)

	if next1 != next2 {
		t.Errorf("Rollback, unexpected diff: %v", (next2 - next1))
	}

	// simulate clock rollback by forwarding elapsed by 10ms
	rollbackDiff := int64(10)
	atomic.AddInt64(&elapsed, rollbackDiff)

	next3 := Next() >> (seqBits + machBits)

	if next3-next1 != 10 {
		t.Errorf("Rollback, unexpected delay: %v", (next3 - next1))
	}
}
