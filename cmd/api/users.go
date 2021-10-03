package main

import (
	"net/http"

	"github.com/kellemNegasi/greenlight/internal/data"
	"github.com/kellemNegasi/greenlight/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request){
	var input struct{
		Name string `json:"name"`
		Email string `json:"email"`
		Password string `json:"password"`
	}


	// parse the request body into the anonymous input struct

	err := app.readJSON(w,r,&input)
	if err !=nil{
		app.badRequestResponse(w,r,err)
		return
	}

	user := &data.User{
		Name : input.Name,
		Email: input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err !=nil{
		app.serveErrorResponse(w,r,err)
		return 
	}

	v:=validator.New()
	if data.ValidateUser(v, user);!v.Valid(){
		app.faildValidationResoponse(w,r,v.Errors)
		return
	}
}