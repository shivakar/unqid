# unqid

**Warning**: This is a pre-alpha software. Things are going to change drastically while we develop towards v1.

Unqid is a distributed unique ID generator library for in-application embedded use.

Components of the ID:

```
| timestamp (40-bits) | machine-id (16-bits) | sequence-number (7-bits)|
```

`timestamp` is calculated since the `epoch` of `Jan. 1, 2021`
`machine-id` is the lower 16 bits of the private ip from the machine's os `pid`
`sequence-number` is monotonically increasing number reset every millisecond

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
import "github.com/shivakar/unqid"

func main() {
    for i := 0; i < 10; i++ {
        fmt.Println(unqid.Next())
    }
}
``` 

## Inspirations
- https://github.com/sony/sonyflake/
- https://github.com/twitter-archive/snowflake