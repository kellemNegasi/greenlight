package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)


func (app *application) routs() http.Handler{
	router:= httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed =http.HandlerFunc(app.methodNotAllowedResoponse)
	// movies related routes 
	router.HandlerFunc(http.MethodGet,"/v1/movies",app.requiredActivatedUser(app.listMovieHandler))
	router.HandlerFunc(http.MethodGet,"/v1/healthcheck",app.requiredActivatedUser(app.healthcheckHandler))
	router.HandlerFunc(http.MethodPost,"/v1/movies",app.requiredActivatedUser(app.createMovieHandler))
	router.HandlerFunc(http.MethodGet,"/v1/movies/:id",app.requiredActivatedUser(app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch,"/v1/movies/:id",app.requiredActivatedUser(app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete,"/v1/movies/:id",app.requiredActivatedUser(app.deleteMovieHandler))
	// routes for user
	router.HandlerFunc(http.MethodPost,"/v1/users",app.registerUserHandler)
	router.HandlerFunc(http.MethodPut,"/v1/users/activated",app.activateUserHandler)
	router.HandlerFunc(http.MethodPost,"/v1/tokens/authentication",app.createAuthenticationTokenHandler)
	return app.recoverPanic(app.rateLimit(app.authenticate(router)))

}