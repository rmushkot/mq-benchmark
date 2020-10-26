package main

import (
	"flag"
	"fmt"
	"runtime"

	"./daemon"
)

const defaultPort = 9000

func main() {
	var port = flag.Int("port", defaultPort, "daemon port")

	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	d, err := daemon.NewDaemon()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Flotilla daemon started on port %d...\n", *port)
	if err := d.Start(*port); err != nil {
		panic(err)
	}
}
