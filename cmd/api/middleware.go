package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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
	// Define a client struct to hold the rate limiter and last seen time for each client.
	type client struct {
		limiter *rate.Limiter
		lastSeen time.Time
	}
	// Declare a mutex and a map to hold the clients' IP addresses and rate limiters.
	var (
		mu sync.Mutex
		clients = make(map[string]*client))
	// launce a background go routine to remove old timer clients from the map every one minute.
	go func(){
		for {
			time.Sleep(time.Minute)
			// lock the  mutex to prevent limiter checks while the routine is on cleaning duties
			mu.Lock()
			for ip,client := range clients{
				if time.Since(client.lastSeen)>time.Minute*3{
					delete(clients,ip) // delete the specific client at ip
				}
			}
			// unlocke the mutex after finishing the cleaning up
			mu.Unlock()
		}
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// conditionaly check the limiter based on the vlaue of limiter enbaled
		if app.config.limiter.enabled{
			ip,_,err := net.SplitHostPort(r.RemoteAddr)
			if err!=nil{
				app.serveErrorResponse(w,r,err)
				return
			}
			// lock this portion of the code from concurency  ------------i.e the code below------
			mu.Lock()
			// check if the ip address already exists in the clients map
			if _,found := clients[ip];!found{
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps),app.config.limiter.burst),
				}
			}
			// update the last seen for the specific client
			clients[ip].lastSeen=time.Now()
	
			// call the allow method of the limiter for the current ip and 
			// if not allowed unlocke the mutex and send 429 i.e Too many requests2
			if !clients[ip].limiter.Allow(){
				mu.Unlock()
				app.rateLimitExceededResponse(w,r)
				return
			}
			mu.Unlock() // unlock the mutex when the limiter check is done
		}
		next.ServeHTTP(w,r)
	})
} 