# bidirpc

[![GoDoc][1]][2] [![MIT licensed][3]][4] [![Build Status][5]][6] [![Coverage Statusd][7]][8]

[1]: https://godoc.org/github.com/zhuyie/bidirpc?status.svg
[2]: https://godoc.org/github.com/zhuyie/bidirpc
[3]: https://img.shields.io/badge/license-MIT-blue.svg
[4]: LICENSE
[5]: https://travis-ci.org/zhuyie/bidirpc.svg?branch=master
[6]: https://travis-ci.org/zhuyie/bidirpc
[7]: https://codecov.io/gh/zhuyie/bidirpc/branch/master/graph/badge.svg
[8]: https://codecov.io/gh/zhuyie/bidirpc

bidirpc is a simple bi-direction RPC library.

## Usage

```go

import (
    "io"
    "log"

    "github.com/zhuyie/bidirpc"
)


var conn io.ReadWriteCloser

// Create a registry, and register your available services, Service follows
// net/rpc semantics
registry := bidirpc.NewRegistry()
registry.Register(&Service{})

// TODO: Establish your connection before passing it to the session

// Create a new session
session, err := bidirpc.NewSession(conn, Yin, registry, 0)
if err != nil {
	log.Fatal(err)
}
// Clean up session resources
defer func() {
	if err := session.Close(); err != nil {
		log.Fatal(err)
	}
}()

// Start the event loop, this is a blocking call, so place it in a goroutine
// if you need to move on.  The call will return when the connection is
// terminated.
if err = session.Serve(); err != nil {
	log.Fatal(err)
}
```
