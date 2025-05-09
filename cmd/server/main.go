package main

import (
	"flag"
	"mini-sgbd/internal/connection"
	"mini-sgbd/internal/pipeline"
)

func main() {
	maxConns := flag.Int("max-connections", 10, "Maximum number of concurrent connections")
	flag.Parse()

	pipeline.StartParseMaster()
	pipeline.StartExecMaster()
	pipeline.StartLogMaster()
	connection.StartServer(":9000", *maxConns)
}
