/*
 * Email Service
 * handle email via basic SMTP
 *
 * Dependencies:
 * github.com/googollee/go-rest
 * 
 * Original 01/26/2016
 * Created by Todd Moses
 */

package main

import (
    "os"
    "fmt"
    "errors"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "net"
    "net/smtp"
    "net/mail"
    "crypto/tls"
    "github.com/googollee/go-rest"
)

const _smtpServer = "email-smtp.us-west-2.amazonaws.com:465"
const _smtpUsername = "AKIAJYT3ZJQYEJCQQGLA"
const _smtpUserPass = "AnoHj+R9cyT4kePGgnwqVwGtxaAe5rDdPB/12wwNkAZz"

const _host = "0.0.0.0"
const _port = 8081
const _version = "0.0.1"
const _copyyear = "(C) 2017"

type emailJson struct {
    From string         `json:"from"`
    To string           `json:"to"`
    Subject string      `json:"subject"`
    Body string         `json:"body"`
}

func sendEmai(fromAddr string, toAddr string, subj string, body string) (err error) {
    
    if len(fromAddr) == 0 || len(toAddr) == 0 || len(subj) == 0 || len(body) == 0 {
        err = errors.New("From, To, Subject, and Body Required")
        return
    }
    
    from := mail.Address{"", fromAddr}
    to   := mail.Address{"", toAddr}
 
    // Setup headers
    headers := make(map[string]string)
    headers["From"] = from.String()
    headers["To"] = to.String()
    headers["Subject"] = subj
 
    // Setup message
    message := ""
    for k,v := range headers {
        message += fmt.Sprintf("%s: %s\r\n", k, v)
    }
    message += "\r\n" + body
 
    // Connect to the SMTP Server
    servername := _smtpServer
 
    host, _, _ := net.SplitHostPort(servername)
 
    auth := smtp.PlainAuth("", _smtpUsername, _smtpUserPass, host)
 
    // TLS config
    tlsconfig := &tls.Config {
        InsecureSkipVerify: true,
        ServerName: host,
    }
 
    // Here is the key, you need to call tls.Dial instead of smtp.Dial
    // for smtp servers running on 465 that require an ssl connection
    // from the very beginning (no starttls)
    conn, err := tls.Dial("tcp", servername, tlsconfig)
    if err != nil {
        err = errors.New("unable to connect to smtp server ssl")
        return
    }
 
    c, err := smtp.NewClient(conn, host)
    if err != nil {
        err = errors.New("unable to connect to smtp server")
        return
    }
 
    // Auth
    if err = c.Auth(auth); err != nil {
        err = errors.New("unable to authorize smtp user")
        return
    }
 
    // To && From
    err = c.Mail(from.Address)
    if err != nil {
        err = errors.New("invalid from address")
        return
    }
    err = c.Rcpt(to.Address) 
    if err != nil {
        err = errors.New("invalid to address")
        return
    }
 
    // Data
    w, err := c.Data()
    if err != nil {
        err = errors.New("unable to create data writer")
        return
    }
 
    _, err = w.Write([]byte(message))
    if err != nil {
        err = errors.New("invalid body or subject")
        return
    }
 
    err = w.Close()
    if err != nil {
        err = errors.New("unable to close smtp connection")
    }
 
    c.Quit()
    
    return
}

func parseEmailJson(rawjson []byte) (to string, from string, subject string, body string, err error) {

	//ensure json is not empty
    if len(rawjson) == 0 {
        err = errors.New("Message is Empty")
        return
    }

    var emailObj emailJson
    
    err = json.Unmarshal(rawjson, &emailObj)
    //format error message
    if err != nil {
        err = errors.New("Invalid Message Format")
        return
    }

    //we got a parsed object now
    to = emailObj.To
    from = emailObj.From
    subject = emailObj.Subject
    body = emailObj.Body

    return
}

func emailSendHandler(w http.ResponseWriter, r *http.Request) {

	//get body
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        //handle and log error
        http.Error(w, "Invalid Request Format", http.StatusBadRequest)
        return
    }
    
    //try to parse json
    to, from, subject, emlBody, err := parseEmailJson(body)
    if err != nil {
        //handle and log error
        http.Error(w, "Invalid Request Body", http.StatusBadRequest)
        return
    }

    err = sendEmai(from,to,subject,emlBody)
    if err != nil {
        //handle and log error
        http.Error(w, "Unable to Send Mail", http.StatusInternalServerError)
        return
    }

    //set headers
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("content-type", "application/json")
    w.Header().Set("X-POWERED-BY", "Tactician Inc")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Headers", "x-api-key, origin, x-requested-with, content-type, accept, referer, user-agent")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
    w.Header().Set("Connection", "close")
    w.WriteHeader(200)
    
    //send response    
    w.Write([]byte("1"))
}

func baseHandler(w http.ResponseWriter, r *http.Request) {

    body := fmt.Sprintf("<h1>Tactician</h1><p>Email Service v%s. Copyright %s by Tactician Inc</p><hr><p>For more info see <a href=\"http://tactician.com\">http://tactician.com</a></p>", _version, _copyyear)
    msgcnt := fmt.Sprintf("<html><head><title>Tactician Status</title><meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\"></head><body>%s</body></html>", body)
    
    fmt.Fprintf(w, "%s", msgcnt)
}

func httpListener(host string, port int) (err error) {
    
    //safety check
    if len(host) == 0 || port <= 0 {
        err = errors.New("Host and Port Required in Configuration")
        return
    }
    
    //implment go-rest
    r := rest.New()
    
    // add router. router must before mime parser because parser need handler's parameter inserted by router.
    r.Use(rest.NewRouter())
    
    //display to console
    fmt.Println("-> Listening:", host, "on Port", port)

    //build address
    addr := fmt.Sprintf("%s:%d", host, port)
    
    //create get handlers
    r.Get("/", baseHandler)
    
    //send an email
    r.Post("/email/send/", emailSendHandler)

    //listen for connections
    http.ListenAndServe(addr, r)
    
    return
}

func main() {

	fmt.Println("================================================")
    fmt.Println("Email Service", _version, _copyyear,"by Tactician Inc")
    fmt.Println("================================================")

    err := httpListener(_host, _port)
    if err != nil {
        //if error occurs exit
        fmt.Println(fmt.Sprintf("ERROR OnStart: %q [Exit]",err))
        os.Exit(1)
    }
    
}