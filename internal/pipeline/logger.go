package pipeline

import (
	"fmt"
	"mini-sgbd/internal/model"
	"sync"
	"time"
)

var logQueue = make([]*model.ParsedCommand, 0)
var logMutex = sync.Mutex{}
var logCond = sync.NewCond(&logMutex)

func StartLogMaster() {
	go func() {
		for {
			logMutex.Lock()
			for len(logQueue) == 0 {
				logCond.Wait()
			}
			cmd := logQueue[0]
			logQueue = logQueue[1:]
			logMutex.Unlock()

			go logWorker(cmd)
		}
	}()
}

func EnqueueLog(cmd *model.ParsedCommand) {
	logMutex.Lock()
	logQueue = append(logQueue, cmd)
	logCond.Signal()
	logMutex.Unlock()
}

func logWorker(cmd *model.ParsedCommand) {
	fmt.Printf("[Logger] Sending result: %s => %s\n", cmd.Raw, cmd.Result)
	time.Sleep(30 * time.Millisecond)
	cmd.Conn.Write([]byte(cmd.Result + "\n"))
}
