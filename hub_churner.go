package main

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/bitly/go-nsq"
	r "github.com/dancannon/gorethink"
	"github.com/dchest/uniuri"
)

type HubEvent struct {
	Type      string `json:"type"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
}

func initChurner(change chan struct{}) {
	cons, err := nsq.NewConsumer("hub", "hub", nsq.NewConfig())
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Unable to consume the hub topic")
	}
	cons.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		log.Print("Handling hub event")

		var ev *HubEvent
		if err := json.Unmarshal(m.Body, &ev); err != nil {
			log.Print("Erroring - invalid body")
			return err
		}

		log.Printf("Type of the hub event is %s", ev.Type)

		switch ev.Type {
		case "onboarding":
			log.Print("Diving into the lock")

			stateLock.Lock()

			log.Print("I'm inside!")

			// Four emails in total
			timers := []*Timer{
				// 1. Welcome to Lavaboom
				&Timer{
					ID:      uniuri.NewLen(uniuri.UUIDLen),
					Time:    time.Now().Add(time.Second * 15),
					Name:    *welcomeName,
					Version: *welcomeVersion,
					Sender:  "hello",
					From:    "Felix from Lavaboom <hello@lavaboom.com>",
					To:      []string{ev.Email},
					Input: map[string]interface{}{
						"first_name": ev.FirstName,
					},
				},
				// 2. Getting started
				&Timer{
					ID:      uniuri.NewLen(uniuri.UUIDLen),
					Time:    time.Now().Add(time.Second * 30),
					Name:    *gettingStartedName,
					Version: *gettingStartedVersion,
					Sender:  "hello",
					From:    "Tine from Lavaboom <hello@lavaboom.com>",
					To:      []string{ev.Email},
					Input: map[string]interface{}{
						"first_name": ev.FirstName,
					},
				},
				// 3. Security information
				&Timer{
					ID:      uniuri.NewLen(uniuri.UUIDLen),
					Time:    time.Now().Add(time.Minute * 2),
					Name:    *securityName,
					Version: *securityVersion,
					Sender:  "hello",
					From:    "Andrei from Lavaboom <hello@lavaboom.com>",
					To:      []string{ev.Email},
					Input: map[string]interface{}{
						"first_name": ev.FirstName,
					},
				},
				// 4. How's it going?
				&Timer{
					ID:      uniuri.NewLen(uniuri.UUIDLen),
					Time:    time.Now().Add(time.Minute * 15),
					Name:    *whatsUpName,
					Version: *whatsUpVersion,
					Sender:  "hello",
					From:    "Lavabot from Lavaboom <hello@lavaboom.com>",
					To:      []string{ev.Email},
					Input: map[string]interface{}{
						"first_name": ev.FirstName,
					},
				},
			}

			state = append(state, timers...)

			if err := r.Db(*rethinkdbDatabase).Table("hub_state").Insert(timers).Exec(session); err != nil {
				log.WithField("error", err.Error()).Error("Unable to insert events into database")
				return err
			}

			// Sort it and ping the worker
			sort.Sort(state)
			change <- struct{}{}
			stateLock.Unlock()
		default:
			return errors.New("Not implemented")
		}

		return nil
	}))
	cons.ConnectToNSQLookupd(*lookupdAddress)
}
