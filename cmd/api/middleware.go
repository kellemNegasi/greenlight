package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/kellemNegasi/greenlight/internal/data"
	"github.com/kellemNegasi/greenlight/internal/validator"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "Close")
				app.serveErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// Define a client struct to hold the rate limiter and last seen time for each client.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	// Declare a mutex and a map to hold the clients' IP addresses and rate limiters.
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	// launce a background go routine to remove old timer clients from the map every one minute.
	go func() {
		for {
			time.Sleep(time.Minute)
			// lock the  mutex to prevent limiter checks while the routine is on cleaning duties
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > time.Minute*3 {
					delete(clients, ip) // delete the specific client at ip
				}
			}
			// unlocke the mutex after finishing the cleaning up
			mu.Unlock()
		}
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// conditionaly check the limiter based on the vlaue of limiter enbaled
		if app.config.limiter.enabled {
			// ip, _, err := net.SplitHostPort(r.RemoteAddr)
			// if err != nil {
				// 	app.serveErrorResponse(w, r, err)
				// 	return
				// }
			
			ip := realip.FromRequest(r)
			// lock this portion of the code from concurency  ------------i.e the code below------
			mu.Lock()
			// check if the ip address already exists in the clients map
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}
			// update the last seen for the specific client
			clients[ip].lastSeen = time.Now()

			// call the allow method of the limiter for the current ip and
			// if not allowed unlocke the mutex and send 429 i.e Too many requests2
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			mu.Unlock() // unlock the mutex when the limiter check is done
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// add the "Vary:Athorization" header to the response. to indicate this may vary based on the
		// value of the autherization header in the request
		w.Header().Add("Vary", "Authorization")
		// get the authorization header

		authorizationHeader := r.Header.Get("Authorization")
		// if there is no authorization header found use the contextSetUser to set anonymous user

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		// otherwise expect the value of the authorization header to be "Bearer"<token>"
		// split into its constituent parts check if the header in the correct format

		headrParts := strings.Split(authorizationHeader, " ")

		// check length and value of the first element
		if len(headrParts) != 2 || headrParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// if that went well lets do the actual work
		token := headrParts[1]
		v := validator.New()
		// check for the validity of the token
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// if the token is valid the retrieve the details of the user associated with token

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serveErrorResponse(w, r, err)
			}
			return
		}
		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)

	})
}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fun := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the user from the context
		user := app.contextGetUser(r)
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fun)
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		permisions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serveErrorResponse(w, r, err)
			return
		}
		if !permisions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}
	return app.requireActivatedUser(fn)
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		// add the Vary:Access-Control-Request-Method
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" && len(app.config.cors.trustedOrigins) != 0 {
			// loop through the list of trusted origins checking to see if the request
			// origin exactly matches one of them
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// check if the request has the HTTP method "OPTIONS" and contains
					// the "Access-Control-Request-Method" header. if it does then we treat it
					// as a prefilight request like so

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.WriteHeader(http.StatusOK)
						return
					}

				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) metrics(next http.Handler) http.Handler {
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResposesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_??s")
	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		totalRequestsReceived.Add(1)
		metrics := httpsnoop.CaptureMetrics(next, w, r)
		// next.ServeHTTP(w,r)
		totalResposesSent.Add(1)

		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}
