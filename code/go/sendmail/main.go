package main

import (
	"os"
	"log"
	"net/smtp"
)

const smtpSendServer ="smtp.gmail.com:587"
const smtpAuthServer ="smtp.gmail.com"

func main() {
    from := os.Getenv("EMAIL")
    if from == "" {
        log.Fatalf("Must set your @gmail address as EMAIL environment variable.\n")
    }
    password := os.Getenv("PWD")
    if password == "" {
        log.Fatalf("Must set your gmail password as PWD environment variable\n")
    }
	err := send(from, password, "essentialgo@mailinator.com", "hello there")
	if err != nil {
		log.Fatalf("smtp error: %s", err)
	}

	log.Print("Email sent, visit https://mailinator.com/v3/index.jsp?zone=public&query=essentialgo#/#inboxpane")
}

func send(from, password, to, body string) error {
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Hello there\n\n" +
		body

    auth := smtp.PlainAuth("", from, password, smtpAuthServer)
	return smtp.SendMail(smtpSendServer, auth, from, []string{to}, []byte(msg))
}