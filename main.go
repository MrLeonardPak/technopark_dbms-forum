package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/MrLeonardPak/technopark_forum-dbms/api"
	"github.com/fasthttp/router"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/valyala/fasthttp"
)

const initialScriptPath = "./db/db.sql"

func main() {

	fRouter := router.New()
	api.DBS = initDB(context.Background(), initialScriptPath)
	api.InitRouters(fRouter.Group("/api"))
	log.Fatal(fasthttp.ListenAndServe(":5000", wrapperHeader(fRouter.Handler)).Error())

}

func wrapperHeader(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Content-Type", "application/json")
		handler(ctx)
	}
}

func initDB(defaultCtx context.Context, initScript string) *pgxpool.Pool {
	connectString := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=Europe/Moscow",
		os.Getenv(("POSTGRES_USER")),
		os.Getenv(("POSTGRES_PASSWORD")),
		os.Getenv(("POSTGRES_HOST")),
		"5432",
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

	sql, err := os.ReadFile(initScript)
	if err != nil {
		log.Fatal("ReadFile error: ", err)
	}

	_, err = pool.Exec(defaultCtx, string(sql))
	if err != nil {
		log.Fatal("Exec error: ", err)
	}

	return pool
}
