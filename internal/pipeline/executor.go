package pipeline

import (
	"fmt"
	"mini-sgbd/internal/db"
	"mini-sgbd/internal/model"
	"sync"
	"time"
)

var execQueue = make([]*model.ParsedCommand, 0)
var execMutex = sync.Mutex{}
var execCond = sync.NewCond(&execMutex)

func StartExecMaster() {
	go func() {
		for {
			execMutex.Lock()
			for len(execQueue) == 0 {
				execCond.Wait()
			}
			cmd := execQueue[0]
			execQueue = execQueue[1:]
			execMutex.Unlock()

			go execWorker(cmd)
		}
	}()
}

func EnqueueExec(cmd *model.ParsedCommand) {
	execMutex.Lock()
	execQueue = append(execQueue, cmd)
	execCond.Signal()
	execMutex.Unlock()
}

func execWorker(cmd *model.ParsedCommand) {
	fmt.Printf("[Executor] Executing: %s\n", cmd.Raw)
	time.Sleep(100 * time.Millisecond)

	switch cmd.Action {
	case "GET":
		db.Mutex.RLock()
		cmd.Result = db.Data[cmd.Key]
		db.Mutex.RUnlock()
	case "SET":
		db.Mutex.Lock()
		db.Data[cmd.Key] = cmd.Value
		db.Mutex.Unlock()
		cmd.Result = "OK"
	default:
		cmd.Result = "ERROR: Unknown Command"
	}

	EnqueueLog(cmd)
}
