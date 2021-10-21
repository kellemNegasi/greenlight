package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/kellemNegasi/greenlight/internal/data"
	"github.com/kellemNegasi/greenlight/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.serveErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.faildValidationResoponse(w, r, v.Errors)
		return
	}

	// if the validation passes
	// then look up the user based on the email if user doesn't exist
	// invalidCredentialResponse should be sent
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serveErrorResponse(w, r, err)

		}
		return
	}
	// if user is found then match the password
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serveErrorResponse(w, r, err)
		return
	}
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}
	// otherwise if the password matches generate a new token for 24hrs and scope authentication
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serveErrorResponse(w, r, err)
		return
	}

	// if that goes well send the token as json
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serveErrorResponse(w, r, err)
		return
	}

}

func (app *application) createPasswordResetTokenHandler(w http.ResponseWriter,r *http.Request){
	// parse and validate users email address

	var input struct{
		Email string `json:"email"`
	}

	err :=app.readJSON(w,r,&input) 
	if err!=nil{
		app.badRequestResponse(w,r,err)
		return
	}
	v:=validator.New()
	if data.ValidateEmail(v,input.Email);!v.Valid(){
		app.faildValidationResoponse(w,r,v.Errors)
		return
	}
	user,err := app.models.Users.GetByEmail(input.Email)
	if err!=nil{
		switch{
		case errors.Is(err,data.ErrRecordNotFound):
			v.AddError("email","no matching email address found")
			app.faildValidationResoponse(w,r,v.Errors)
		default:
			app.serveErrorResponse(w,r,err)
		}
		return
	}
	if !user.Activated{
		v.AddError("email","user account must be activated")
		app.faildValidationResoponse(w,r,v.Errors)
		return
	}

	// otherwise create a new password reset token with a 45 minutes expiry time.
	token,err := app.models.Tokens.New(user.ID,45*time.Minute,data.ScopePasswordReset)
	if err!=nil{
		app.serveErrorResponse(w,r,err)
		return
	}

	app.background(func() {
		data:= map[string]interface{}{
			"passwordResetToken":token.Plaintext,
		}
		err = app.mailer.Send(user.Email,"token_password_reset.tmpl",data)
		if err!=nil{
			app.logger.PrintError(err,nil)
		}
	})

	env := envelope{"message":"an email will be sent to you containing password reset instructions"}
	err = app.writeJSON(w,http.StatusAccepted,env,nil)
	if err!=nil{
		app.serveErrorResponse(w,r,err)
	}
}

func (app *application) createActivationTokenHandler(w http.ResponseWriter,r *http.Request){
	var input struct{
		Email string `json:"email`

	}

	err := app.readJSON(w,r,&input)
	if err!=nil{
		app.badRequestResponse(w,r,err)
		return
	}
	v:= validator.New()
	if data.ValidateEmail(v,input.Email);!v.Valid(){
		app.faildValidationResoponse(w,r,v.Errors)
		return
	}

	// if email is valid then, try to retrieve the user associated with email

	user,err := app.models.Users.GetByEmail(input.Email)
	if err!=nil{
		switch {
		case errors.Is(err,data.ErrRecordNotFound):
			v.AddError("email","no matching email address found")
			app.faildValidationResoponse(w,r,v.Errors)
		default:
			app.serveErrorResponse(w,r,err)
		}
		return
	}
	if user.Activated{
		v.AddError("email","user is already activated")
		app.faildValidationResoponse(w,r,v.Errors)
		return
	}
	// otherwise proceed with activation and generate activation token
	token,err := app.models.Tokens.New(user.ID,24*time.Hour,data.ScopeActivation)
	if err!=nil{
		app.serveErrorResponse(w,r,err)
		return
	}

	app.background(func() {
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			}
		err:=app.mailer.Send(user.Email,"token_activation.tmpl",data)
		if err!=nil{
			app.logger.PrintError(err,nil)
		}
	})
	env := envelope{"message": "an email will be sent to you containing activation instructions"}
	err= app.writeJSON(w,http.StatusAccepted,env,nil)
	if err!=nil{
		app.serveErrorResponse(w,r,err)
	}
}
