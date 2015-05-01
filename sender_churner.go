package main

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/bitly/go-nsq"
	"github.com/blang/semver"
	"github.com/lavab/api/client"
	"github.com/lavab/api/routes"
)

type SenderEvent struct {
	Name    string      `gorethink:"name"`
	Version string      `gorethink:"version"`
	To      []string    `gorethink:"to"`
	Input   interface{} `gorethink:"input"`
}

func initSender(username, password string) {
	api, err := client.New(*apiURL, 0)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Unable to connect to the Lavaboom API")
	}

	token, err := api.CreateToken(&routes.TokensCreateRequest{
		Type:     "auth",
		Username: username,
		Password: password,
	})
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Unable to sign into Lavaboom's API")
	}

	api.Headers["Authorization"] = "Bearer " + token.ID

	cons, err := nsq.NewConsumer("sender_"+username, "sender", nsq.NewConfig())
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Unable to consume the hub topic")
	}

	cons.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		var ev *SenderEvent
		if err := json.Unmarshal(m.Body, &ev); err != nil {
			return err
		}

		// Parse the version
		version, err := semver.Parse(ev.Version)
		if err != nil {
			return err
		}

		templateLock.RLock()
		// Check if we have such template
		if x, ok := templates[ev.Name]; !ok || len(x) > 0 {
			templateLock.RUnlock()
			return errors.New("No such template")
		}

		// Match the version
		found := -1
		for i := len(templateVersions[ev.Name]) - 1; i >= 0; i-- {
			v2 := templateVersions[ev.Name][i]
			if version.Major == v2.Major {
				found = i
				break
			}
		}

		// Get the template
		template := templates[ev.Name][templateVersions[ev.Name][found].String()]

		// Execute the subject
		subject := &bytes.Buffer{}
		if err := template.SubjectTpl.Execute(subject, ev.Input); err != nil {
			return err
		}

		// Execute the body
		body := &bytes.Buffer{}
		if err := template.BodyTpl.Execute(body, ev.Input); err != nil {
			return err
		}

		// Send the email
		resp, err := api.CreateEmail(&routes.EmailsCreateRequest{
			Kind:        "raw",
			To:          ev.To,
			Body:        body.String(),
			Subject:     subject.String(),
			ContentType: "text/html",
		})
		if err != nil {
			return err
		}

		// Log some debug info
		log.Printf("Sent a \"%s\" email to %v - %v", subject.String(), ev.To, resp)

		// We're done!
		return nil
	}))

	cons.ConnectToNSQLookupd(*lookupdAddress)
}
