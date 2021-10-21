package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/kellemNegasi/greenlight/internal/data"
	"github.com/kellemNegasi/greenlight/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// parse the request body into the anonymous input struct

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serveErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.faildValidationResoponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.faildValidationResoponse(w, r, v.Errors)
		default:
			app.serveErrorResponse(w, r, err)
		}
		return
	}
	//add the read permission by default

	err = app.models.Permissions.AddForUser(user.ID, "movies:read")
	if err != nil {
		app.serveErrorResponse(w, r, err)
		return
	}
	// generate a new token for the user
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeActivation)
	// send the email using the mailer package i.e calling the send method on the mailer object
	// to avoid overhead send the mail in a background own go routine

	app.background(func() {
		// let's crate a new map to hold the data
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
			"userName":        user.Name,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil) // use this instead of app.ServeErrorResponse to avoid redudant response
		}

	})
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serveErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.faildValidationResoponse(w, r, v.Errors)
		return
	}
	// get the details of the user associated with token using the GetForToken() method
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.faildValidationResoponse(w, r, v.Errors)
		default:
			app.serveErrorResponse(w, r, err)
		}
		return
	}
	// if that goes well then update the user's activation status
	user.Activated = true
	// then update the usre

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConfilict):
			app.editConflictResponse(w, r, err)
		default:
			app.serveErrorResponse(w, r, err)
		}
		return
	}
	// if everything went succesfully ... the delete all activation tokens for the user
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serveErrorResponse(w, r, err)
		return
	}

	// send the updated user details back to the client
	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serveErrorResponse(w, r, err)
	}
}

func (app *application) updateUserPasswordHandler(w http.ResponseWriter,r *http.Request){
	var input struct{
		Password string `json:"password"`
		TokenPlaintext string `json:"token"`
	}
	err:= app.readJSON(w,r,&input)
	if err!=nil{
		app.badRequestResponse(w,r,err)
		return
	}
	// validate the password and the token
	v:= validator.New()
	data.ValidatePasswordPlaintext(v,input.Password)
	data.ValidateTokenPlaintext(v,input.TokenPlaintext)
	if !v.Valid(){
		app.faildValidationResoponse(w,r,v.Errors)
		return
	}

	// if it is valid then retrieve the corrosponding user
	user,err := app.models.Users.GetForToken(data.ScopePasswordReset,input.TokenPlaintext)
	if err!=nil{
		switch{
		case errors.Is(err,data.ErrRecordNotFound):
			v.AddError("token","invalid or expired reset token")
			app.faildValidationResoponse(w,r,v.Errors)
		default:
			app.serveErrorResponse(w,r,err)
		}
		return
	}
	// if the user exists which means the token is valid, update the password for the user
	err = user.Password.Set(input.Password)
	if err!=nil{
		app.serveErrorResponse(w,r,err)
		return
	}

	// then save the update user, i.e with reset password in database
	err = app.models.Users.Update(user)
	if err!=nil{
		switch{
		case errors.Is(err,data.ErrEditConfilict):
			app.editConflictResponse(w,r,err)
		default:
			app.serveErrorResponse(w,r,err)
		}
		return 
	}
	// if that goes well then send a confirmatin message to the user
	env := envelope{"message":"your password was successfully reset"}
	err = app.writeJSON(w,http.StatusOK,env,nil)
	if err !=nil{
		app.serveErrorResponse(w,r,err)
		return
	}

}