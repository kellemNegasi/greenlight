package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/kellemNegasi/greenlight/internal/validator"
)

const (
	ScopeActivation = "activation"
)

// create token struct to represent token type

type Token struct {
	Plaintext 	string
	Hash		[]byte
	UserID		int64
	Expiry		time.Time
	Scope		string
}

// token model for database interaction wrapper
type TokenModel struct{
	DB *sql.DB
}

func ValidateTokenPlaintext(v *validator.Validator ,tokenPlaintext string){
	v.Check(tokenPlaintext!="","token","must be provided")
	v.Check(len(tokenPlaintext)==26,"token","must be 26 bytes long")
}



func generateToken(userID int64, ttl time.Duration,scope string)(*Token,error){
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope: scope,
	}
// initialize a zero valuesd byte slice with a length of 16 bytes.
	randomBytes :=make([]byte,16)
	// fill the randomBytes slice with random bytes using the Read function
	_,err := rand.Read(randomBytes)
	if err!=nil{
		return nil,err
	}

	// encode the byte slice with a base 32-encoded string and assign
	// it the plaintext feild. this is the token sent to the user

	token.Plaintext=base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	// generate SHA-256 hash of the plaintext 
	// this will be stored in the hash feild of hte our database table

	hash:=sha256.Sum256([]byte(token.Plaintext))
	// change it to slice and assign to the Hash field of the token object
	token.Hash = hash[:]
	return token,nil
}

// database related methods on the TokenModel

func (m TokenModel) New(userID int64,ttl time.Duration,scope string)(*Token,error){
	token,err := generateToken(userID,ttl,scope)
	if err!=nil{
		return nil,err
	}

	err = m.Insert(token)
	return token,err
}

func (m TokenModel) Insert(token *Token) error{
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)`
	args:= []interface{}{token.Hash,token.UserID,token.Expiry,token.Scope}
	ctx,cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	_,err := m.DB.ExecContext(ctx,query,args...)
	return err
}

func (m TokenModel) DeleteAllForUser(scope string, userID int64) error{
	query:= `
		DELETE FROM tokens
		WHERE scope = $1 AND user_id = $2
	`
	arges := []interface{}{scope,userID}
	ctx,cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()

	_,err:= m.DB.ExecContext(ctx,query,arges...)
	return err
}