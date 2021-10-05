package mailer

import (
	"bytes"
	"embed"
	"text/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"

var templateFs embed.FS

// define mailer struct 
type Mailer struct{
	dialer *mail.Dialer 			// used to connect to an smtp servr
	sender string 				// contains sender inforamtions 

}

func New(host string, port int,username,password,sender string) Mailer{
	dialer := mail.NewDialer(host,port,username,password)
	dialer.Timeout = 5*time.Second
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

func (m *Mailer) Send(recipient,templateFile string,data interface{}) error{
	tmpl,err :=template.New("email").ParseFS(templateFs,"templates/"+templateFile)
	if err!=nil{
		return err
	}
	// execute the named template subject passing the dynamic datea and storing it 
	// in the subject buffer
	subject:= new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject,"subject",data)
	if err!=nil{
		return err
	}
	// do the same for the named plain body template
	plainBody :=new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody,"plainBody",data)
	if err!=nil{
		return err
	}

	// and follow the same pattern for the "htmlBody" template inside the template file
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody,"htmlBody",data)
	if err!=nil{
		return err
	}

	// initialize a new mail.Message and set the necessary feilds
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())
	// call the DialAndSend on the dialer feild of the Mailer to send
	// this opens a connections to the SMTP server sends the message and closes the connection
	// if there is a time out, "dial tcp:i/o timeout" will be  returned

	// on hindsight let's make the sender try 3 times before aborting

	for i:=1;i<=3;i++{
		err = m.dialer.DialAndSend(msg)
		// if everythind is fine return nil
		if err==nil{
			return nil
		}
		// if it didn't send try again after 500 millisecond i.e half a second
		time.Sleep(500*time.Millisecond)
	}
	return err
}




