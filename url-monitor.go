package main

import (
	"fmt"
	// "io/ioutil"
	"crypto/tls"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"time"
)

type SmtpServer struct {
	Address  string
	Username string
	Password string
}

func (ss *SmtpServer) SendEmail(email string, subj string, body string) {

	// this is very good
	// https://gist.github.com/chrisgillis/10888032
	// thanks very much
	fmt.Printf("send email to %s : %s | %s\n", email, subj, body)

	from := mail.Address{"", ss.Username}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = email
	// fmt.Printf("%s\n", to.String())
	// return
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	servername := ss.Address

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", ss.Username, ss.Password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		log.Panic(err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Panic(err)
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		log.Panic(err)
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		log.Panic(err)
	}

	if err = c.Rcpt(email); err != nil {
		log.Panic(err)
	}

	// Data
	w, err := c.Data()
	if err != nil {
		log.Panic(err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		log.Panic(err)
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
	}

	c.Quit()
}

// http-monitor "http://example.com/" "admin@my.com"
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <url> <email>\n", os.Args[0])
		return
	}
	url := os.Args[1]
	email := os.Args[2]
	ss := readConfig("smtp.json")
	for {
		checkUrl(ss, url, email)
		time.Sleep(time.Minute)
	}
}

func readConfig(file_name string) SmtpServer {
	f, err := os.Open(file_name)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	// b , err := ioutil.ReadAll(f)
	// fmt.Printf("%s\n", string(b))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	d := json.NewDecoder(f)
	o := SmtpServer{}
	err = d.Decode(&o)
	if err != nil {
		log.Fatal(err)
	}
	return o
}

func checkUrl(ss SmtpServer, url string, email string) {
	// fmt.Printf("checkUrl %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	var s string
	fmt.Printf("%s %d\n", url, resp.StatusCode)
	if resp.StatusCode != 200 {
		s = fmt.Sprintf("%s is down, http status %d", url, resp.StatusCode)
		ss.SendEmail(email, s, s)
		return
	}
	// fmt.Printf("%v\n", resp.Header)
	// fmt.Printf("%v\n", resp.Header["Content-Length"])
	if cllist, ok := resp.Header["Content-Length"]; ok {
		if len(cllist) == 1 {
			cls := cllist[0]
			cl, err := strconv.Atoi(cls)
			if err != nil {
				s = fmt.Sprintf("%s Content-Length %s to int error", url, cls)
				ss.SendEmail(email, s, s)
			}
			if cl < 14000 {
				s = fmt.Sprintf("%s Content-Length %s, too short", url, cls)
				ss.SendEmail(email, s, s)
			}
		} else {
			s = fmt.Sprintf("%s Content-Length #%d", url, len(cllist))
			ss.SendEmail(email, s, s)
		}
	} else {
		s = fmt.Sprintf("%s has no Content-Length", url)
		ss.SendEmail(email, s, s)
	}
}
