package main

import (
	"mini-sgbd/internal/connection"
	"mini-sgbd/internal/pipeline"
)

func main() {
	pipeline.StartParseMaster()
	pipeline.StartExecMaster()
	pipeline.StartLogMaster()
	connection.StartServer(":9000")
}
