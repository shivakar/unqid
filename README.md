# unqid

![example branch parameter](https://github.com/shivakar/unqid/actions/workflows/go.yml/badge.svg?branch=main) [![Coverage Status](https://coveralls.io/repos/github/shivakar/unqid/badge.svg?branch=main)](https://coveralls.io/github/shivakar/unqid?branch=main) [![Go Reference](https://pkg.go.dev/badge/github.com/shivakar/unqid.svg)](https://pkg.go.dev/github.com/shivakar/unqid)

**Warning**: This is a pre-alpha software. Things are going to change drastically while we develop towards v1.

Unqid is a distributed unique ID generator library for in-application embedded use.

Components of the ID:

```
| timestamp (40-bits) | machine-id (16-bits) | sequence-number (7-bits)|
```

`timestamp` is calculated since the `epoch` of `Jan. 1, 2021`
`machine-id` is the lower 16 bits of the private ip from the machine's os `pid`
`sequence-number` is monotonically increasing number reset to 0 every millisecond

Due to the component sizes:
- The lifetime of id generation is ~34 years (2^40 - 1 milliseconds from custom epoch)
- concurrently used from ~65k hosts (2^16-1)
- generate 128k ids/sec per instance


## When to use (and not use)

Use this library if:
- You will generate less than 128 ids/ms (or 128k ids/sec) per instance of unqid
- You run only one instance of your app per machine or use containers with networking
- You're confident that instances of your app don't run on subnets with overlapping lower 16-bits of the network
- You're ok with your application having a theoretical lifetime limit of 34 years.

## Design Goals

- No external dependencies (this might be relaxed in later versions)
- Zero configuration, i.e., import library and start generating ids (this might be relaxed in later versions)
- Roughly time sortable
- Fast id generation
- Protect against clock rollbacks
- Protect against accidental overflow of `sequence-number` component

## Installation

```
go get github.com/shivakar/unqid
```

## Usage

```
package main

import (
	"fmt"

	"github.com/shivakar/unqid"
)

func main() {
	for i := 0; i < 10; i++ {
		fmt.Println(unqid.Next())
	}
}


// Output
// 375314704840587009
// 375314704840587010
// 375314704840587011
// 375314704840587012
// 375314704840587013
// 375314704840587014
// 375314704840587015
// 375314704840587016
// 375314704840587017
// 375314704840587018
``` 


### Testing

```
$ go test -v -cover -covermode=atomic -bench .
=== RUN   TestInitPanics
=== RUN   TestInitPanics/machineID_error
=== RUN   TestInitPanics/machineID_zero
=== RUN   TestInitPanics/max_time_exceeded
--- PASS: TestInitPanics (0.00s)
    --- PASS: TestInitPanics/machineID_error (0.00s)
    --- PASS: TestInitPanics/machineID_zero (0.00s)
    --- PASS: TestInitPanics/max_time_exceeded (0.00s)
=== RUN   Test_machineID
=== RUN   Test_machineID/valid1
=== RUN   Test_machineID/loopback
=== RUN   Test_machineID/no-private
=== RUN   Test_machineID/valid2
=== RUN   Test_machineID/invalid
--- PASS: Test_machineID (0.00s)
    --- PASS: Test_machineID/valid1 (0.00s)
    --- PASS: Test_machineID/loopback (0.00s)
    --- PASS: Test_machineID/no-private (0.00s)
    --- PASS: Test_machineID/valid2 (0.00s)
    --- PASS: Test_machineID/invalid (0.00s)
=== RUN   TestNext
=== RUN   TestNext/maxTime
=== RUN   TestNext/maxSequence
=== RUN   TestNext/valid1
--- PASS: TestNext (0.00s)
    --- PASS: TestNext/maxTime (0.00s)
    --- PASS: TestNext/maxSequence (0.00s)
    --- PASS: TestNext/valid1 (0.00s)
=== RUN   TestNextRollback
--- PASS: TestNextRollback (0.01s)
=== RUN   ExampleNext
--- PASS: ExampleNext (0.25s)
goos: darwin
goarch: arm64
pkg: github.com/shivakar/unqid
BenchmarkUnqidID
BenchmarkUnqidID-8        154314              7814 ns/op
PASS
coverage: 100.0% of statements
ok      github.com/shivakar/unqid       1.639s
```

That's ~128k ids/sec.

## Inspirations
- https://github.com/sony/sonyflake/
- https://github.com/twitter-archive/snowflake