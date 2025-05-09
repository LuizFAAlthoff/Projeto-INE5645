package pipeline

import (
	"fmt"
	"mini-sgbd/internal/config"
	"mini-sgbd/internal/model"
	"sync"
	"time"
)

var logQueue = make([]*model.ParsedCommand, 0) //fila de log
var logMutex = sync.Mutex{}                    //mutex para a fila de log
var logCond = sync.NewCond(&logMutex)          //condição para a fila de log
var logWg sync.WaitGroup                       //grupo de espera para os workers de log

func StartLogMaster() {
	for i := 0; i < config.LogWorkers; i++ { //inicia o logger com o número de workers especificado em config.LogWorkers
		logWg.Add(1)   //adiciona um worker ao grupo de espera
		go logWorker() //inicia o worker
	}
}

func logWorker() { //função para o logger
	defer logWg.Done() //finaliza o worker quando a função terminar
	for {
		logMutex.Lock()          //bloqueia o acesso ao logQueue
		for len(logQueue) == 0 { //verifica se a fila de log está vazia
			logCond.Wait() //espera a fila de log ter um comando
		}
		cmd := logQueue[0]      //pega o primeiro comando da fila
		logQueue = logQueue[1:] //remove o primeiro comando da fila
		logMutex.Unlock()       //libera o acesso ao logQueue

		fmt.Printf("[Logger] Sending result: %s => %s\n", cmd.Raw, cmd.Result) //printa o comando sendo logado
		time.Sleep(30 * time.Millisecond)                                      //espera 30ms
		cmd.Conn.Write([]byte(cmd.Result + "\n"))                              //escreve o resultado do comando na conexão
	}
}

func EnqueueLog(cmd *model.ParsedCommand) { //função para enviar o comando para o logger
	logMutex.Lock()                  //bloqueia o acesso ao logQueue
	logQueue = append(logQueue, cmd) //adiciona o comando à fila de log
	logCond.Signal()                 //sinaliza que há um comando na fila de log
	logMutex.Unlock()                //libera o acesso ao logQueue
}

func StopLogWorkers() { //função para parar os workers do logger
	logWg.Wait() //espera todos os workers terminarem
}
