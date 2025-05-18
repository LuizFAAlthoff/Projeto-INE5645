package pipeline

import (
	"fmt"
	"mini-sgbd/internal/config" //importação do pacote config, que contém as configurações do servidor (número de workers, etc)
	"mini-sgbd/internal/model"
	"strings"
	"sync"
	"time"
)

var parseQueue = make([]*model.ParsedCommand, 0)
var parseMutex = sync.Mutex{}
var parseCond = sync.NewCond(&parseMutex)
var parseWg sync.WaitGroup

func StartParseMaster() {
	//inicia o parser com o número de workers especificado em config.ParseWorkers
	for i := 0; i < config.ParseWorkers; i++ { //loop para iniciar os workers
		parseWg.Add(1)   //adiciona um worker ao grupo de espera
		go parseWorker() //inicia o worker
	}
}

func parseWorker() { //função para o parser
	defer parseWg.Done() //finaliza o worker quando a função terminar
	for {
		parseMutex.Lock()          //bloqueia o acesso ao parseQueue
		for len(parseQueue) == 0 { //verifica se a fila de parse está vazia
			parseCond.Wait() //espera a fila de parse ter um comando   !!! IMPORTANTE !!!
		}
		cmd := parseQueue[0]        //pega o primeiro comando da fila
		parseQueue = parseQueue[1:] //remove o primeiro comando da fila
		parseMutex.Unlock()         //libera o acesso ao parseQueue

		fmt.Printf("[Parser] Parsing: %s\n", cmd.Raw) //printa o comando sendo parseado
		time.Sleep(50 * time.Millisecond)             //espera 50ms

		parts := strings.Fields(cmd.Raw) //divide o comando em partes
		if len(parts) >= 2 {             //verifica se o comando tem pelo menos 2 partes
			cmd.Action = strings.ToUpper(parts[0])      //define a ação do comando
			cmd.Key = parts[1]                          //define a chave do comando
			if cmd.Action == "SET" && len(parts) == 3 { //verifica se a ação é SET e se tem 3 partes
				cmd.Value = parts[2] //define o valor do comando
			}
		}

		EnqueueExec(cmd) //envia o comando para o executor
	}
}

func EnqueueParse(cmd *model.ParsedCommand) { //função para enviar o comando para o parser
	parseMutex.Lock()                    //bloqueia o acesso ao parseQueue
	parseQueue = append(parseQueue, cmd) //adiciona o comando à fila de parse
	parseCond.Signal()                   //sinaliza que há um comando na fila de parse   !!! IMPORTANTE !!!
	parseMutex.Unlock()                  //libera o acesso ao parseQueue
}

func StopParseWorkers() { //função para parar os workers do parser
	parseWg.Wait() //espera todos os workers terminarem
}
