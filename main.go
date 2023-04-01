package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/MrLeonardPak/technopark_forum-dbms/api"
	"github.com/MrLeonardPak/technopark_forum-dbms/middlewares"
	"github.com/fasthttp/router"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/valyala/fasthttp"
)

func main() {
	fRouter := router.New()

	middlewares.InitPrometheus(fRouter)

	api.DBS = initDB(context.Background())
	api.InitRouters(fRouter.Group("/api"))

	log.Fatal(fasthttp.ListenAndServe(":"+os.Getenv(("SERVER_PORT")), fRouter.Handler).Error())
}

func initDB(defaultCtx context.Context) *pgxpool.Pool {
	connectString := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=Europe/Moscow",
		os.Getenv(("POSTGRES_USER")),
		os.Getenv(("POSTGRES_PASSWORD")),
		os.Getenv(("POSTGRES_HOST")),
		os.Getenv(("POSTGRES_PORT")),
		os.Getenv(("POSTGRES_DB")),
	)
	fmt.Println(connectString)
	connectConfig, err := pgxpool.ParseConfig(connectString)

	if err != nil {
		log.Fatal("ParseConfig error: ", err)
	}

	connectConfig.MaxConns = 128
	connectConfig.MaxConnLifetime = time.Minute
	connectConfig.MaxConnIdleTime = time.Second * 5

	pool, err := pgxpool.ConnectConfig(defaultCtx, connectConfig)
	if err != nil {
		log.Fatal("ConnectConfig error: ", err)
	}

	if err = pool.Ping(defaultCtx); err != nil {
		log.Fatal("Ping error: ", err)
	}

	return pool
}
