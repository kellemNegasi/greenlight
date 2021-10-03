package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/kellemNegasi/greenlight/internal/validator"
)

type envelope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64,error){
	params := httprouter.ParamsFromContext(r.Context())

	id,err:= strconv.ParseInt(params.ByName("id"),10,64)
	if err!=nil || id<1{
		return 0,errors.New("invalid id parameter")
	}
	return id,nil
}

func (app *application) writeJSON(w http.ResponseWriter,status int,data interface{},headers http.Header) error{
	js,err :=json.MarshalIndent(data,"","\t")
	if err!=nil{
		return err
	}
	js = append(js,'\n')
	for key,value:=range headers{
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter,r *http.Request,dst interface{}) error{
	maxBytes:= 1_048_576
	r.Body = http.MaxBytesReader(w,r.Body,int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err:= dec.Decode(dst)
	if err!=nil{
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch{
		case errors.As(err,&syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)",syntaxError.Offset)
		case errors.Is(err,io.ErrUnexpectedEOF):
			return errors.New("body contains incorrect JSON type for field")
		case errors.As(err,&unmarshalTypeError):
			if unmarshalTypeError.Field!=""{
				return fmt.Errorf("body contains incorrect JSON type for field %q",unmarshalTypeError.Field)
	
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)",unmarshalTypeError.Offset)
		case errors.Is(err,io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err,&invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	err = dec.Decode(&struct{}{})
	if err!=io.EOF{
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}
func(app *application) readString(qs url.Values,key string,defaultValue string) string{
	// get the target string using the given key
	s:=qs.Get(key)

	// if value is not given return the default value
	if s==""{
		return defaultValue
	}
	return s
}

func(app *application) readCSV(qs url.Values,key string,defaultValue []string) []string{
	csv := qs.Get(key)
	if csv==""{
		return defaultValue
	}
	return strings.Split(csv,",")
}

func (app *application) readInt(qs url.Values,key string,defaultValue int,v *validator.Validator) int{
	s:=qs.Get(key)
	if s==""{
		return defaultValue
	}
	// converted the string to intger
	i,err := strconv.Atoi(s)
	// check for error
	if err!=nil{
		v.AddError(key,"must be an intger value")
		return defaultValue
	}
	//  return the valid converted integer
	return i
}

func (app *application) background(fn func()){
	// increament wait group counter to signify a new go routine
	app.wg.Add(1)
	go func(){
		// defered revory function
		defer func(){
			if err:=recover();err!=nil{
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		fn()
	}()
}