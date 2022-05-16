package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	d := net.Dialer{ //nolint:exhaustivestruct
		Timeout:   time.Second,
		KeepAlive: time.Minute,
	}

	conn, err := d.DialContext(ctx, "tcp", "[::1]:9001")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(io.Copy(os.Stdout, conn))
}
