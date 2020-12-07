package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/sekaiichi/gosql/cmd/app"
	"github.com/sekaiichi/gosql/pkg/customers"

	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	host := "0.0.0.0"
	port := "9999"
	dsn := "postgres://app:pass@localhost:5432/db"

	if err := execute(host, port, dsn); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(host string, port string, dsn string) (err error) {
	connectCtx, _ := context.WithTimeout(context.Background(), time.Second * 5)
	pool, err := pgxpool.Connect(connectCtx, dsn)
	if err != nil {
		log.Println(err)
		return
	}
	defer pool.Close()

	mux := http.NewServeMux()
	customersSvc := customers.NewService(pool)
	server := app.NewServer(mux, customersSvc)
	server.Init()

	srv := &http.Server{
		Addr: net.JoinHostPort(host, port),
		Handler: server,
	}
	return srv.ListenAndServe()
}