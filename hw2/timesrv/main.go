package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	cfg := net.ListenConfig{ //nolint:exhaustivestruct
		KeepAlive: time.Minute,
	}

	l, err := cfg.Listen(ctx, "tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}

	log.Println("im started!")

	go func() {
		for {
			var conn net.Conn
			conn, err = l.Accept()

			if err != nil {
				log.Println(err)

				return
			}

			wg.Add(1)

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
		}
	}
}
