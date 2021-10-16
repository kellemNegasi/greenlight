package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kellemNegasi/greenlight/internal/data"
	"github.com/kellemNegasi/greenlight/internal/jsonlog"
	"github.com/kellemNegasi/greenlight/internal/mailer"
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
	// add a new limiter struct to hole the limiter configurations
	limiter struct{
		rps float64 // requests per second
		burst int // burst value
		enabled bool // 
	}	
	// add smtp related configs
	smtp struct {
		host 		string
		port 		int
		username 	string
		password 	string
		sender 		string
		}
	cors struct {
		trustedOrigins []string
	}	
}
type application struct{
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg sync.WaitGroup
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
	
	// Create command line flags to read the setting values into the config struct.
	// Notice that we use true as the default for the 'enabled' setting?
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// command line flags for smtp settings from mailtrap

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "8cac89e168c738", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "cce1dcf8487a2d", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.kellemnegasi.net>", "SMTP sender")
	
	// use the Flag.func() function to process the -cors-trusted-origns command line flag
	flag.Func("cors-trusted-origins","Trusted CORS origins (space separated)",func(val string)error{
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})
	
	// parese the recieved values


	flag.Parse()
	// logger:= log.New(os.Stdout,"",log.Ldate|log.Ltime)
	logger :=jsonlog.New(os.Stdout,jsonlog.LevelInfo)
	db,err :=openDB(cfg)
	if err!=nil{

		logger.PrintFatal(err,nil)
	}

	defer db.Close()
	logger.PrintInfo("database connection pool established!",nil)
	
	app:= &application{
		config: cfg,
		logger:  logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
	// call the serve() method to startup the server.
	err=app.serve()	
	if err!=nil{
		logger.PrintFatal(err,nil)
	}
}

// function to create DB connection 

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
