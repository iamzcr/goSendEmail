package email

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"html/template"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

type (
	Email struct {
		username string
		password string
		host     string
		port     int
		ssl      bool
	}
	Object struct {
		To      []string
		Header  map[string]string
		Content string
	}
	HTML struct {
		body []byte
	}
	config struct {
		Username string `toml:"username"`
		Password string `toml:"password"`
		Host     string `toml:"host"`
		Ssl      bool   `toml:"ssl"`
		Port     int    `toml:"port"`
	}
)

var (
	email   *Email
	auth    smtp.Auth
	senders chan *Object
	once    sync.Once
	conf    config
)

func LoadEmailConfig(dir, fileName, fileType string) (err error) {
	viper.AddConfigPath(dir)
	viper.SetConfigName(fileName)
	viper.SetConfigType(fileType)
	err = viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("config error: %s \n", err))
	}
	conf.Username = viper.GetString("email.username")
	conf.Password = viper.GetString("email.password")
	conf.Host = viper.GetString("email.host")
	conf.Port = viper.GetInt("email.port")
	conf.Ssl = viper.GetBool("email.ssl")
	fmt.Println(conf)
	return err
}
func StartEmail() {
	once.Do(func() {
		email = &Email{
			username: conf.Username,
			password: conf.Password,
			host:     conf.Host,
			ssl:      conf.Ssl,
			port: func() int {
				if conf.Port < 1 {
					return 25
				}
				return conf.Port
			}(),
		}
		auth = smtp.PlainAuth("", email.username, email.password, email.host)
		senders = make(chan *Object, 4096)
		go task()
	})
}

func ParseHHTML(path string, data interface{}) *HTML {
	html := new(HTML)
	html.body = make([]byte, 0)
	parse, err := template.ParseFiles(path)
	if err != nil {
		panic(fmt.Errorf("ParseHHTML error: %s \n", err))
		return nil
	}

	buffer := bytes.NewBuffer(html.body)
	if err = parse.Execute(buffer, data); err != nil {
		panic(fmt.Errorf("ParseHHTML error: %s \n", err))
		return nil
	}

	html.body = buffer.Bytes()
	return html
}

func NewSender(subject, content string, to ...string) *Object {
	object := &Object{
		Content: content,
		Header:  make(map[string]string),
		To:      to,
	}
	object.writeHeader("Subject", subject).
		writeHeader("From", email.username).
		writeHeader("To", strings.Join(to, ";")).
		writeHeader("Mime-Version", "1.0").
		writeHeader("Date", time.Now().String())

	return object
}

func NewHTMLSender(subject string, html *HTML, to ...string) *Object {
	object := NewSender(subject, string(html.body), to...)
	object.writeHeader("Content-Type", "text/html;chartset=UTF-8")
	return object
}

func (object *Object) writeHeader(key, value string) *Object {
	object.Header[key] = value
	return object
}

func (object *Object) convertToBody() []byte {
	headers := ""
	for key, value := range object.Header {
		headers += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	return []byte(headers + "\r\n" + object.Content)
}

func Send(sender *Object) {
	senders <- sender
}

func task() {
	for {
		select {
		case object := <-senders:
			if email.ssl {
				sendBySsl(object)
			} else {
				err := smtp.SendMail(fmt.Sprintf("%s:%d", email.host, email.port), auth, email.username, object.To, object.convertToBody())
				if err != nil {
					panic(fmt.Errorf("task error: %s \n", err))
				}
			}
		}
	}
}
