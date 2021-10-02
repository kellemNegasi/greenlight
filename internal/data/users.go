package data

import (
	"errors"
	"time"

	"github.com/kellemNegasi/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type User struct{
	ID 		  	 int64		 `json:"id"`
	CreatedAt	 time.Time	 `json:"created_at"` 
	Name 		 string		 `json:"name"`
	Email	 	 string  	 `json:"email"`
	Password	 password  	 `json:"password"`
	Activated	 bool		 `json:"-"`
	Version 	 int  		 `json:"version"`

}

type password struct{
	plaintext *string // should be a pointer to distinguish between empty string and no string provided
	hash 	  []byte
}

func (p *password) Set(plaintTextPassword string) error{
	hash,err := bcrypt.GenerateFromPassword([]byte(plaintTextPassword),12)
	if err!=nil{
		return err
	}
	p.plaintext = &plaintTextPassword
	p.hash = hash
	return  nil
}

func (p *password) Matches(plaintTextPassword string) (bool,error){
	err:=bcrypt.CompareHashAndPassword(p.hash,[]byte(plaintTextPassword))
	if err!=nil{
		switch{
		case errors.Is(err,bcrypt.ErrMismatchedHashAndPassword):
			return false,nil
		default:
			return false,err
		}
	}
	return true,nil
}

func ValidateEmail(v *validator.Validator,email string){
	v.Check(email!="","email","must be provided")
	v.Check(validator.Matches(email,validator.EmailRX),"email","must be a valid email addres")
}
func ValidatePasswordPlaintext(v *validator.Validator,password string){
	v.Check(password!="","password","must be provided")
	v.Check(len(password)>=8,"password","must be at least 8 bytes long")
	v.Check(len(password)<=72,"passowrd","must not be more than 72 bytes long")
}

func ValidateUser( v *validator.Validator, user *User){
	v.Check(user.Name!= "","name","must be given")
	v.Check(len(user.Name)<=500,"name","must be less thatn 500 bytes")
	
	// check the email here
	ValidateEmail(v,user.Email)
	if user.Password.plaintext!=nil{
		ValidatePasswordPlaintext(v,*user.Password.plaintext)
	}
	// let's do some sanity check if hash is empty due to some logic mistake
	if user.Password.hash==nil{
		panic("password hash is empty for some reason")
	}

}
