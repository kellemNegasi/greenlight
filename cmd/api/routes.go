package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)


func (app *application) routs() http.Handler{
	router:= httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed =http.HandlerFunc(app.methodNotAllowedResoponse)
	// movies related routes 
	router.HandlerFunc(http.MethodGet,"/v1/healthcheck",app.healthcheckHandler)


	router.HandlerFunc(http.MethodGet,"/v1/movies",app.requirePermission("movies:read",app.listMovieHandler))
	router.HandlerFunc(http.MethodPost,"/v1/movies",app.requirePermission("movies:write",app.createMovieHandler))
	router.HandlerFunc(http.MethodGet,"/v1/movies/:id",app.requirePermission("movies:read",app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch,"/v1/movies/:id",app.requirePermission("movies:write",app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete,"/v1/movies/:id",app.requirePermission("movies:write",app.deleteMovieHandler))
	// routes for user
	router.HandlerFunc(http.MethodPost,"/v1/users",app.registerUserHandler)
	router.HandlerFunc(http.MethodPut,"/v1/users/activated",app.activateUserHandler)
	router.HandlerFunc(http.MethodPost,"/v1/tokens/authentication",app.createAuthenticationTokenHandler)
	// routes for expvar metrics

	router.Handler(http.MethodGet,"/v1/metrics",expvar.Handler())
	return  app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))

}