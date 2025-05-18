package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	serverAddr        = "localhost:9000"
	numClients        = 15 // Quantidade de clientes simultâneos
	commandsPerClient = 5  // Quantidade de comandos que cada cliente enviará
)

var actions = []string{"SET", "GET"}

func auto_command() {
	var wg sync.WaitGroup
	wg.Add(numClients)

	for i := 0; i < numClients; i++ {
		go func(id int) {
			defer wg.Done()
			simulateClient(id)
		}(i)
	}

	wg.Wait()
	fmt.Println("Todos os clientes terminaram.")
}

func simulateClient(clientID int) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("[Cliente %d] Erro ao conectar: %v\n", clientID, err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for i := 0; i < commandsPerClient; i++ {
		action := actions[rand.Intn(len(actions))]
		key := fmt.Sprintf("chave%d", rand.Intn(5)) // Algumas chaves podem se repetir para forçar concorrência
		var command string

		if action == "SET" {
			value := fmt.Sprintf("valor%d_cliente%d", rand.Intn(100), clientID)
			command = fmt.Sprintf("SET %s %s", key, value)
		} else { // GET
			command = fmt.Sprintf("GET %s", key)
		}

		fmt.Fprintf(conn, "%s\n", command)

		// Lê a resposta
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("[Cliente %d] Erro ao ler resposta: %v\n", clientID, err)
			return
		}
		fmt.Printf("[Cliente %d] Comando: %s | Resposta: %s", clientID, command, response)

		// Pequeno atraso aleatório para simular comportamento real
		time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(200)))
	}
}

func manual_command() {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Conectado ao Mini SGBD. Digite comandos como GET chave ou SET chave valor")
	reader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(conn)

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao ler comando:", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "exit" || line == "quit" {
			fmt.Println("Encerrando cliente.")
			break
		}

		_, err = conn.Write([]byte(line + "\n"))
		if err != nil {
			fmt.Println("Erro ao enviar comando:", err)
			return
		}

		// Aguarda resposta do servidor
		response, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao ler resposta:", err)
			return
		}
		fmt.Printf("[Servidor] %s", response)
	}
}

func main() {
	mode := flag.String("mode", "auto", "Mode: manual or auto")
	flag.Parse()

	if *mode == "manual" {
		manual_command()
	} else if *mode == "auto" {
		auto_command()
	} else {
		fmt.Println("Invalid mode. Use --mode=manual or --mode=auto")
	}
}
