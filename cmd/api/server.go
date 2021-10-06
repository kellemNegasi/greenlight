package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error{
// declare the http sever with the following settings
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d",app.config.port),
		Handler: app.routs(),
		IdleTimeout : time.Minute,
		ReadTimeout: 10*time.Second,
		WriteTimeout: 30*time.Second,}
	// create a shutdownerr channel to recieve any error from the graceful shutdown
	shutdownError := make(chan error)
	// spin up a background go routine to catch singlas
	go func(){
		// create a channel which carries os.Signal. i.e quit channel
		quit := make(chan os.Signal,1)
		// listen for incoming SIGINT and SIGTERM signals using signal.notify()
		// any other signals will not be caught by our channel and should retain theri original behaviour.
		signal.Notify(quit,syscall.SIGINT,syscall.SIGTERM)
		// read the signal from the channel.
		//------ N.B this code blockes until the signal is recieved------
		s:= <-quit

		//---------------------------------------------

		// log the recieving of the signal 
		app.logger.PrintInfo("shuting down server",map[string]string{
			"signal":s.String(),
		})
		// create a context with 5 seconds timeout
		ctx,cancel := context.WithTimeout(context.Background(),time.Second*5)
		defer cancel()
		// call shutdown() on our server by passing the context
		err := srv.Shutdown(ctx)
		if err!=nil{
			shutdownError<- err
		}
		// log message that we are waiting for background tasks
		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})
		// call the wait on the waitGroup
		app.wg.Wait()
		shutdownError <- nil

	}()
	app.logger.PrintInfo("starting server on",map[string]string{
		"addr":srv.Addr,
		"env":app.config.env,
	})

	err:= srv.ListenAndServe()
	// calling shutdown causes Listend and server to return the http.ErrServerClosed error.
	// we are looking for this error. If not just return it to the caller.
	if !errors.Is(err,http.ErrServerClosed){
		return err
	}
	// otherwise we wait to get the return value from the channel
	// if the value is an error something must have happend
	err = <- shutdownError
	if err !=nil{
		return err
	}
	// now that all the above errors are passed it means 
	//at this point it is sure that the graceful shutdown has completed successfully

	app.logger.PrintInfo("stopped server ",map[string]string{
		"addr":srv.Addr,
	})
	return nil
}