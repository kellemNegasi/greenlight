package main

import (
	"fmt"
	"net/http"
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
	app.logger.PrintInfo("starting server on",map[string]string{
		"addr":srv.Addr,
		"env":app.config.env,
	})

	return srv.ListenAndServe()
}