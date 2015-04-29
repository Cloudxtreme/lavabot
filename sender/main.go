package main

import (
	"bytes"
	"html/template"
	"log"
	"time"

	"github.com/lavab/api/client"
	"github.com/lavab/api/routes"
	"github.com/namsral/flag"
)

var (
	configFlag = flag.String("config", "", "Config file to read")
	apiURL     = flag.String("api_url", "https://api.lavaboom.com", "Path to the Lavaboom API")
	username   = flag.String("username", "", "Username of the Lavaboom account")
	password   = flag.String("password", "", "Either password or a SHA256 hash of the account's password")
	privateKey = flag.String("private_key", "", "Path to the private key")
)

type tplInput struct {
	FirstName string
}

func main() {
	flag.Parse()

	var (
		welcomeTpl  = template.Must(template.New("welcome").Parse(welcomeTemplate))
		startedTpl  = template.Must(template.New("started").Parse(startedTemplate))
		securityTpl = template.Must(template.New("security").Parse(securityTemplate))
		whatsupTpl  = template.Must(template.New("whatsup").Parse(whatsupTemplate))
	)

	client, err := client.New(*apiURL, 0)
	if err != nil {
		log.Fatal(err)
	}

	token, err := client.CreateToken(&routes.TokensCreateRequest{
		Type:     "auth",
		Username: *username,
		Password: *password,
	})
	if err != nil {
		log.Fatal(err)
	}

	client.Headers["Authorization"] = "Bearer " + token.ID

	const target = "piotr@zduniak.net"

	input := tplInput{
		FirstName: "Piotr",
	}

	go func() {
		time.Sleep(welcomeDelay)
		log.Print("Sending welcome")

		output := &bytes.Buffer{}
		log.Print(welcomeTpl.Execute(output, input))

		resp, err := client.CreateEmail(&routes.EmailsCreateRequest{
			Kind:        "raw",
			To:          []string{target},
			Body:        output.String(),
			Subject:     welcomeSubject,
			ContentType: "text/html",
		})
		if err != nil {
			log.Print(err)
		} else {
			log.Printf("Sent! %v", resp)
		}
	}()

	go func() {
		time.Sleep(startedDelay)
		log.Print("Sending started")

		output := &bytes.Buffer{}
		log.Print(startedTpl.Execute(output, input))

		resp, err := client.CreateEmail(&routes.EmailsCreateRequest{
			Kind:        "raw",
			To:          []string{target},
			Body:        output.String(),
			Subject:     startedSubject,
			ContentType: "text/html",
		})
		if err != nil {
			log.Print(err)
		} else {
			log.Printf("Sent! %v", resp)
		}
	}()

	go func() {
		time.Sleep(securityDelay)
		log.Print("Sending security")

		output := &bytes.Buffer{}
		log.Print(securityTpl.Execute(output, input))

		resp, err := client.CreateEmail(&routes.EmailsCreateRequest{
			Kind:        "raw",
			To:          []string{target},
			Body:        output.String(),
			Subject:     securitySubject,
			ContentType: "text/html",
		})
		if err != nil {
			log.Print(err)
		} else {
			log.Printf("Sent! %v", resp)
		}
	}()

	go func() {
		time.Sleep(whatsupDelay)
		log.Print("Sending whatsup")

		output := &bytes.Buffer{}
		log.Print(whatsupTpl.Execute(output, input))

		resp, err := client.CreateEmail(&routes.EmailsCreateRequest{
			Kind:        "raw",
			To:          []string{target},
			Body:        output.String(),
			Subject:     whatsupSubject,
			ContentType: "text/html",
		})
		if err != nil {
			log.Print(err)
		} else {
			log.Printf("Sent! %v", resp)
		}
	}()

	select {}
}
