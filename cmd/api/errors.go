package main

import (
	"fmt"
	"net/http"
)

func (app *application) loggError(r *http.Request, err error){
	app.logger.Println(err)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}){
	env:=envelope{"error":message}
	err:= app.writeJSON(w,status,env,nil)
	if err!=nil{
		app.loggError(r,err)
		w.WriteHeader(500)

	}
}

func (app *application) serveErrorResponse(w http.ResponseWriter,r *http.Request, err error){
	app.loggError(r,err)
	message:= "the sever encountered a problem and couldn't process your request"
	app.errorResponse(w,r,http.StatusNotFound,message)
}

func (app *application) notFoundResponse(w http.ResponseWriter,r *http.Request){
	message:= "the requested resource could not be found"
	app.errorResponse(w,r,http.StatusNotFound,message)
}

func (app *application) methodNotAllowedResoponse(w http.ResponseWriter,r *http.Request){
	message:= fmt.Sprintf("the %s method is not supported for this resource",r.Method)
	app.errorResponse(w,r,http.StatusMethodNotAllowed,message)
}
func (app *application) badRequestResponse(w http.ResponseWriter,r *http.Request,err error){
	app.errorResponse(w,r,http.StatusBadRequest,err.Error())
}

func (app *application) faildValidationResoponse(w http.ResponseWriter,r *http.Request,errors map[string]string){
	app.errorResponse(w,r,http.StatusUnprocessableEntity,errors)
}

func (app *application) editConflictResponse(w http.ResponseWriter,r *http.Request,err error){
	message:="unable to update the record due to edit conflict, please try agin."
	app.errorResponse(w,r,http.StatusConflict,message)
}