package pipeline

import (
	"fmt"
	"mini-sgbd/internal/config"
	"mini-sgbd/internal/db"
	"mini-sgbd/internal/model"
	"sync"
	"time"
)

var execQueue = make([]*model.ParsedCommand, 0) //fila de execução
var execMutex = sync.Mutex{}                    //mutex para a fila de execução
var execCond = sync.NewCond(&execMutex)         //condição para a fila de execução
var execWg sync.WaitGroup                       //grupo de espera para os workers de execução

func StartExecMaster() {
	for i := 0; i < config.ExecWorkers; i++ { //inicia o executor com o número de workers especificado em config.ExecWorkers
		execWg.Add(1)   //adiciona um worker ao grupo de espera
		go execWorker() //inicia o worker
	}
}

func execWorker() { //função para o executor
	defer execWg.Done() //finaliza o worker quando a função terminar
	for {
		execMutex.Lock()          //bloqueia o acesso ao execQueue
		for len(execQueue) == 0 { //verifica se a fila de execução está vazia
			execCond.Wait() //espera a fila de execução ter um comando
		}
		cmd := execQueue[0]       //pega o primeiro comando da fila
		execQueue = execQueue[1:] //remove o primeiro comando da fila
		execMutex.Unlock()        //libera o acesso ao execQueue

		fmt.Printf("[Executor] Executing: %s\n", cmd.Raw) //printa o comando sendo executado
		time.Sleep(100 * time.Millisecond)                //espera 100ms

		// dorme um tempo aleatorio entre 1 e 15 segundos
		// time.Sleep(time.Duration(rand.Intn(15)+1) * time.Second)

		switch cmd.Action {
		case "GET":
			db.Mutex.RLock()              //bloqueia o acesso ao banco de dados  !!! IMPORTANTE !!!
			cmd.Result = db.Data[cmd.Key] //define o resultado do comando
			db.Mutex.RUnlock()            //libera o acesso ao banco de dados   !!! IMPORTANTE !!!
		case "SET":
			db.Mutex.Lock()              //bloqueia o acesso ao banco de dados
			db.Data[cmd.Key] = cmd.Value //define o valor do comando
			db.Mutex.Unlock()            //libera o acesso ao banco de dados
			cmd.Result = "OK"            //define o resultado do comando
		default:
			cmd.Result = "ERROR: Unknown Command"
		}

		EnqueueLog(cmd) //envia o comando para o logger
	}
}

func EnqueueExec(cmd *model.ParsedCommand) { //função para enviar o comando para o executor
	execMutex.Lock()                   //bloqueia o acesso ao execQueue
	execQueue = append(execQueue, cmd) //adiciona o comando à fila de execução
	execCond.Signal()                  //sinaliza que há um comando na fila de execução
	execMutex.Unlock()                 //libera o acesso ao execQueue
}

func StopExecWorkers() { //função para parar os workers do executor
	execWg.Wait() //espera todos os workers terminarem
}
