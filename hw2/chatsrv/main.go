package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

type client chan<- string

const maxNumber = 1000

var (
	entering   = make(chan client)
	leaving    = make(chan client)
	messages   = make(chan string)
	gameResult = -1
)

func main() {
	rand.Seed(time.Now().UnixNano())

	listener, err := net.Listen("tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("server started at localhost:8001")

	go broadcaster()

	log.Println("Type 'game' and press enter to start new math game.")

	go startNewGame()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)

			continue
		}

		go handleConn(conn)
	}
}

func startNewGame() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() != "game" {
			continue
		}

		message := generateNewGameSet()
		messages <- generateNewGameSet()

		log.Println(message)
	}
}

func generateNewGameSet() string {
	op1 := rand.Intn(maxNumber) //nolint:gosec
	op2 := rand.Intn(maxNumber) //nolint:gosec
	gameResult = op1 + op2

	return fmt.Sprintf("New math game: %d + %d = ?", op1, op2)
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	ch <- "You are " + who
	messages <- who + " has arrived"
	entering <- ch

	log.Println(who + " has arrived")

	input := bufio.NewScanner(conn)
	for input.Scan() {
		message := input.Text()
		messages <- who + ": " + input.Text()

		if gameResult == -1 {
			continue
		}

		cliResult, err := strconv.Atoi(message)
		if err != nil {
			continue
		}

		if cliResult == gameResult {
			resultMessage := "The game is over! The winner is " + who
			messages <- resultMessage

			gameResult = -1

			log.Println(resultMessage)
		}
	}
	leaving <- ch
	messages <- who + " has left"

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
