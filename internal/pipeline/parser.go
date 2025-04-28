package pipeline

import (
	"fmt"
	"mini-sgbd/internal/model"
	"strings"
	"sync"
	"time"
)

var parseQueue = make([]*model.ParsedCommand, 0)
var parseMutex = sync.Mutex{}
var parseCond = sync.NewCond(&parseMutex)

func StartParseMaster() {
	go func() {
		for {
			parseMutex.Lock()
			for len(parseQueue) == 0 {
				parseCond.Wait()
			}
			cmd := parseQueue[0]
			parseQueue = parseQueue[1:]
			parseMutex.Unlock()

			go parseWorker(cmd)
		}
	}()
}

func EnqueueParse(cmd *model.ParsedCommand) {
	parseMutex.Lock()
	parseQueue = append(parseQueue, cmd)
	parseCond.Signal()
	parseMutex.Unlock()
}

func parseWorker(cmd *model.ParsedCommand) {
	fmt.Printf("[Parser] Parsing: %s\n", cmd.Raw)
	time.Sleep(50 * time.Millisecond)

	parts := strings.Fields(cmd.Raw)
	if len(parts) >= 2 {
		cmd.Action = strings.ToUpper(parts[0])
		cmd.Key = parts[1]
		if cmd.Action == "SET" && len(parts) == 3 {
			cmd.Value = parts[2]
		}
	}

	EnqueueExec(cmd)
}
