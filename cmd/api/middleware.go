package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		defer func(){
			if err:=recover();err!=nil{
				w.Header().Set("Connection","Close")
				app.serveErrorResponse(w,r,fmt.Errorf("%s",err))
			}
		}()
		next.ServeHTTP(w,r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler{
	// Declare a mutex and a map to hold the clients' IP addresses and rate limiters.
	var (
		mu sync.Mutex
		clients = make(map[string]*rate.Limiter)
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip,_,err := net.SplitHostPort(r.RemoteAddr)
		if err!=nil{
			app.serveErrorResponse(w,r,err)
			return
		}
		// lock this portion of the code from concurency  ------------i.e the code below------
		mu.Lock()
		// check if the ip address already exists in the clients map
		if _,found := clients[ip];!found{
			clients[ip] = rate.NewLimiter(2,4) // if the ip doesnt exist initialize a new limiter and add it to the map
		}

		// call the allow method of the limiter for the current ip and 
		// if not allowed unlocke the mutex and send 429 i.e Too many requests2
		if !clients[ip].Allow(){
			mu.Unlock()
			app.rateLimitExceededResponse(w,r)
			return
		}
		mu.Unlock() // unlock the mutex when the limiter check is done
		next.ServeHTTP(w,r)
	})
} 