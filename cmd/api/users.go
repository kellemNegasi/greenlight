package main

import (
	"errors"
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

	err = app.models.Users.Insert(user)
	
	if err!=nil{
		switch{
		case errors.Is(err,data.ErrDuplicateEmail):
			v.AddError("email","a user with this email address already exists")
			app.faildValidationResoponse(w,r,v.Errors)
		default:
			app.serveErrorResponse(w,r,err)
		}
		return 
	}

	// send the email using the mailer package i.e calling the send method on the mailer object
	err = app.mailer.Send(user.Email,"user_welcome.tmpl",user)
	if err!=nil{
		app.serveErrorResponse(w,r,err)
		return
	}


	err = app.writeJSON(w,http.StatusCreated,envelope{"user":user},nil)
	if err!=nil{
		app.serveErrorResponse(w,r,err)
	}
}