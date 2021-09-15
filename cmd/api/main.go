package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"
type config struct{
	port int
	env string
	db struct{
		dsn string
	}		
}
type application struct{
	config config
	logger *log.Logger
}
func main(){
	var cfg config
	
	flag.IntVar(&cfg.port, "port", 4000, "API server port ")
	flag.StringVar(&cfg.env,"env","development","Environment (development|staging|production")
	flag.StringVar(&cfg.db.dsn,"db-dsn","postgres://greenlight:pa55word@localhost/greenlight","PostgressSQL DSN")

	flag.Parse()
	logger:= log.New(os.Stdout,"",log.Ldate|log.Ltime)
	
	db,err :=openDB(cfg)
	if err!=nil{
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Printf("database connection pool established!")
	
	app:= &application{
		config: cfg,
		logger:  logger,
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d",cfg.port),
		Handler: app.routs(),
		IdleTimeout : time.Minute,
		ReadTimeout: 10*time.Second,
		WriteTimeout: 30*time.Second,
	}
	logger.Printf("starting %s server on %d ",cfg.env,cfg.port)
	err=srv.ListenAndServe()	
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB,error){
	db,err:=sql.Open("postgres",cfg.db.dsn)
	if err!=nil{
		return nil,err
	}

	ctx,cancel := context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err!=nil{
		return nil,err
	}

	return db,nil
}
