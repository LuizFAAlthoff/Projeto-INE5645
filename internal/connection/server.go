package connection

import (
	"bufio"
	"fmt"
	"mini-sgbd/internal/model"
	"mini-sgbd/internal/pipeline"
	"net"
	"sync"
)

var (
	activeConnections int
	connMutex         sync.Mutex
	maxConnections    int
	connSemaphore     chan struct{}
)

func StartServer(address string, maxConns int) { //função para iniciar o servidor, recebe a porta e o número máximo de conexões
	maxConnections = maxConns                     //define o número máximo de conexões
	connSemaphore = make(chan struct{}, maxConns) //cria a fila de conexões

	ln, err := net.Listen("tcp", address) //inicia o servidor na porta especificada
	if err != nil {
		panic(err) //encerra o programa caso ocorra um erro na inicialização do servidor
	}
	fmt.Printf("Server listening on %s (max connections: %d)\n", address, maxConns) //printa a mensagem de que o servidor está escutando na porta e o número máximo de conexões
	for {
		conn, err := ln.Accept() //aceita as conexões dos clientes
		if err != nil {
			fmt.Printf("[Server] Connection error: %v\n", err) //printa a mensagem de erro caso ocorra um erro na aceitação da conexão
			continue
		}

		select { //seleciona a conexão
		case connSemaphore <- struct{}{}: //adiciona a conexão na fila de conexões
			connMutex.Lock()                                                                                             //bloqueia o acesso ao contador de conexões
			activeConnections++                                                                                          //incrementa o contador de conexões
			connMutex.Unlock()                                                                                           //libera o acesso ao contador de conexões
			fmt.Printf("[Server] New client connected (active connections: %d/%d)\n", activeConnections, maxConnections) //printa a mensagem de que uma nova conexão foi aceita
			go handleClient(conn)                                                                                        //chama a função para lidar com a conexão
		default:
			fmt.Printf("[Server] Connection rejected - maximum connections reached (%d)\n", maxConnections) //printa a mensagem de que a conexão foi rejeitada porque o número máximo de conexões foi atingido
			conn.Close()                                                                                    //fecha a conexão
		}
	}
}

func handleClient(conn net.Conn) { //função para lidar com as conexões dos clientes, recebe uma conexão e fecha a conexão quando o cliente se desconecta
	defer func() {
		conn.Close()        //fecha a conexão
		<-connSemaphore     //libera o espaço na fila de conexões
		connMutex.Lock()    //bloqueia o acesso ao contador de conexões
		activeConnections-- //decrementa o contador de conexões
		connMutex.Unlock()  //libera o acesso ao contador de conexões
		fmt.Printf("[Server] Client disconnected (active connections: %d/%d)\n", activeConnections, maxConnections)
	}()

	scanner := bufio.NewScanner(conn) //cria um scanner para ler a entrada do cliente
	for scanner.Scan() {              //lê a entrada do cliente
		line := scanner.Text()    //pega a linha lida
		submitCommand(conn, line) //envia a linha para o parser
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("[Client] Error: %v\n", err) //printa a mensagem de erro caso ocorra um erro na leitura da entrada do cliente
	}
}

func submitCommand(conn net.Conn, raw string) { //função para enviar o comando para o parser
	cmd := &model.ParsedCommand{ //cria um comando parseado
		Conn: conn, //define a conexão
		Raw:  raw,  //define a linha lida
	}
	pipeline.EnqueueParse(cmd) //envia o comando para o parser
}
