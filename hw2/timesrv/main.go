package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

var messages = make(chan string)
var clients []net.Conn

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	cfg := net.ListenConfig{ //nolint:exhaustivestruct
		KeepAlive: time.Minute,
	}

	l, err := cfg.Listen(ctx, "tcp", ":9001")
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}

	log.Println("im started!")

	log.Println("Type text and press enter for send bulk message")

	go sendBulkMessage()

	go func() {
		for {
			var conn net.Conn
			conn, err = l.Accept()

			if err != nil {
				log.Println(err)

				return
			}

			wg.Add(1)

			clients = append(clients, conn)

			go handleConn(ctx, conn, wg)
		}
	}()

	<-ctx.Done()

	log.Println("done")

	err = l.Close()
	if err != nil {
		return
	}

	wg.Wait()
	log.Println("exit")
}

func sendBulkMessage() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		messages <- scanner.Text()
	}
}

func handleConn(ctx context.Context, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()
	// каждую 1 секунду отправлять клиентам текущее время сервера
	tck := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-tck.C:
			_, err := fmt.Fprintf(conn, "now: %s\n", t)
			if err != nil {
				return
			}
		case msg := <-messages:
			for _, cliConn := range clients {
				_, err := fmt.Fprintf(cliConn, "Admin message: %s\n", msg)
				if err != nil {
					return
				}
			}
		}
	}
}
