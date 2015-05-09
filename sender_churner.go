package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/mail"
	"strings"

	"github.com/bitly/go-nsq"
	"github.com/blang/semver"
	"github.com/lavab/api/client"
	"github.com/lavab/api/routes"
	"github.com/lavab/mailer/shared"
	man "github.com/lavab/pgp-manifest-go"
	"golang.org/x/crypto/openpgp"
)

type SenderEvent struct {
	Name    string      `gorethink:"name"`
	From    string      `gorethink:"from"`
	Version string      `gorethink:"version"`
	To      []string    `gorethink:"to"`
	Input   interface{} `gorethink:"input"`
}

func initSender(username, password string) {
	// Open the key file
	/*keyFile, err := os.Open(keyPath)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Unable to open the private key file")
		return
	}

	// Parse the key
	keyring, err := openpgp.ReadArmoredKeyRing(r)*/
	// jk, what am i smoking

	api, err := client.New(*apiURL, 0)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Unable to connect to the Lavaboom API")
		return
	}

	token, err := api.CreateToken(&routes.TokensCreateRequest{
		Type:     "auth",
		Username: username,
		Password: password,
	})
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Unable to sign into Lavaboom's API")
		return
	}

	api.Headers["Authorization"] = "Bearer " + token.ID

	cons, err := nsq.NewConsumer("sender_"+username, "sender", nsq.NewConfig())
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Unable to consume the hub topic")
	}

	cons.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		log.Print("Handling sender event")

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
		if x, ok := templates[ev.Name]; !ok || len(x) == 0 {
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

		// Send unencrypted
		if !strings.HasSuffix(ev.To[0], "@lavaboom.com") {
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
		} else {
			// Get user's key
			key, err := api.GetKey(ev.To[0])
			if err != nil {
				m.Finish()
				m.DisableAutoResponse()
				return err
			}

			// Parse the key
			keyReader := strings.NewReader(key.Key)
			keyring, err := openpgp.ReadArmoredKeyRing(keyReader)
			if err != nil {
				return err
			}

			// Hash the subject
			subjectHash := sha256.Sum256(subject.Bytes())

			// Hash the body
			bodyHash := sha256.Sum256(body.Bytes())

			// Parse the from address
			from, err := mail.ParseAddress(ev.From)
			if err != nil {
				return err
			}

			// Manifest definition
			manifest := &man.Manifest{
				Version: semver.Version{1, 0, 0, nil, nil},
				From:    from,
				To: []*mail.Address{
					{
						Address: ev.To[0],
					},
				},
				Subject: subject.String(),
				Parts: []*man.Part{
					{
						Hash:        hex.EncodeToString(bodyHash[:]),
						ID:          "body",
						ContentType: "text/html",
						Size:        body.Len(),
					},
				},
			}

			// Encrypt the body
			ebody, err := shared.EncryptAndArmor(body.Bytes(), keyring)
			if err != nil {
				return err
			}

			// Generate the manifest
			sman, err := man.Write(manifest)
			if err != nil {
				return err
			}
			eman, err := shared.EncryptAndArmor(sman, keyring)
			if err != nil {
				return err
			}

			// Send the email
			resp, err := api.CreateEmail(&routes.EmailsCreateRequest{
				Kind:        "manifest",
				From:        ev.From,
				To:          ev.To,
				Body:        string(ebody),
				Manifest:    string(eman),
				SubjectHash: hex.EncodeToString(subjectHash[:]),
			})
			if err != nil {
				return err
			}

			// Log some debug info
			log.Printf("Sent an encrypted \"%s\" email to %v - %v", subject.String(), ev.To, resp)

			// We're done!
			return nil
		}
	}))

	cons.ConnectToNSQLookupd(*lookupdAddress)
}
