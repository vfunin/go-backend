package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
	names    = map[string]string{}
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("server started at localhost:8001")

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)

			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	var cliName string

	ch := make(chan string)
	go clientWriter(conn, ch)

	input := bufio.NewScanner(conn)
	cliAddr := conn.RemoteAddr().String()

	_, ok := names[cliAddr]
	if !ok {
		input.Scan()
		names[cliAddr] = input.Text()
	}

	cliName = names[cliAddr]

	ch <- "You are " + cliName
	messages <- cliName + " has arrived"
	entering <- ch

	log.Println(cliName + "[" + cliAddr + "]" + " has arrived")

	for input.Scan() {
		messages <- cliName + ": " + input.Text()
	}
	leaving <- ch
	messages <- cliName + " has left"

	err := conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		_, err := fmt.Fprintln(conn, msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func broadcaster() {
	clients := make(map[client]bool)

	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}
