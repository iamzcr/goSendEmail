package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

func tcpConn() (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", email.host, email.port), nil)
	if err != nil {
		panic(fmt.Errorf("tcpConn error: %s \n", err))
		return nil, err
	}
	return smtp.NewClient(conn, email.host)
}

func sendBySsl(sender *Object) {
	c, err := tcpConn()
	if !catch(err) {
		defer c.Close()
	}

	if ok, _ := c.Extension("AUTH"); ok {
		if err = c.Auth(auth); !catch(err) {
			panic(fmt.Errorf("AUTH error: %s \n", err))
			return
		}
	}

	err = c.Mail(email.username)
	if !catch(err) {
		panic(fmt.Errorf("username error: %s \n", err))
		return
	}

	for _, addr := range sender.To {
		err = c.Rcpt(addr)
		if !catch(err) {
			panic(fmt.Errorf("addr error: %s \n", err))
			return
		}
	}

	w, err := c.Data()
	if !catch(err) {
		panic(fmt.Errorf("w error: %s \n", err))
		return
	}

	_, err = w.Write(sender.convertToBody())
	if !catch(err) {
		panic(fmt.Errorf("Write error: %s \n", err))
		return
	}

	err = w.Close()
	if !catch(err) {
		panic(fmt.Errorf("Close error: %s \n", err))
		return
	}

	err = c.Quit()
	if err != nil {
		panic(fmt.Errorf("Quit error: %s \n", err))
		return
	}
}

func catch(err error) bool {
	if err != nil {
		panic(err)
		return false
	}
	return true
}
