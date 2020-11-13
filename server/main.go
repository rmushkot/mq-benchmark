package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/rmushkot/mq-benchmark/server/daemon"
)

const defaultPort = 9500

func main() {
	var port = flag.Int("port", defaultPort, "daemon port")

	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	d, err := daemon.NewDaemon()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Benchmark daemon started on port %d...\n", *port)
	if err := d.Start(*port); err != nil {
		panic(err)
	}
}
