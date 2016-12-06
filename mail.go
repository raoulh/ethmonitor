package main

import (
	"log"
	"net/smtp"
)

func sendEmail(subject, message string) {
	from := config.EmailNotif
	pass := config.EmailPass
	to := config.EmailNotif

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		message

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
	}
}
