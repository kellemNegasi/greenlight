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
		dsn 		 string
		maxOpenConns int
		maxIdleConns int
		maxIldeTime  string
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
	flag.StringVar(&cfg.db.dsn,"db-dsn",os.Getenv("GREENLIGHT_DB_DSN"),"PostgressSQL DSN")

	// Read connection pool setting from command line arguments
	flag.IntVar(&cfg.db.maxOpenConns,"db-max-open-conns",25,"PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns,"db-max-idle-conns",25,"PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIldeTime,"db-max-idle-time","15m","PostgreSQL max connection idle time")
	
	// parese the recieved values


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
	// set the connection pool settings from cfg
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	// first parse the value of the maxIdleTime to int before giving it to the setter

	duration,err := time.ParseDuration(cfg.db.maxIldeTime)
	if err!=nil{
		return nil,err
	}
	// set the max idle time
	db.SetConnMaxIdleTime(duration)



	ctx,cancel := context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err!=nil{
		return nil,err
	}

	return db,nil
}
