package main

import (
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/bitly/go-nsq"
	r "github.com/dancannon/gorethink"
)

var (
	state     State
	stateLock sync.Mutex
)

func initHub(change chan struct{}) {
	// Create a new producer
	prod, err := nsq.NewProducer(*nsqdAddress, nsq.NewConfig())
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Unable to connect to NSQd")
	}

	// Load the hub state from RethinkDB
	cursor, err := r.Db(*rethinkdbDatabase).Table("hub_state").OrderBy(r.OrderByOpts{
		Index: "time",
	}).Run(session)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Unable to query for hub state")
	}

	var result State
	if err := cursor.All(&result); err != nil {
		log.WithField("error", err.Error()).Fatal("Unable to parse hub state query")
	}

	sort.Sort(result)
	state = result

	for {
		// TODO: Use MultiPublish
		stateLock.Lock()
		timersToDelete := []int{}
		for id, timer := range state {
			if timer.Time.Before(time.Now()) {
				// Encode the sender event
				body, err := json.Marshal(&SenderEvent{
					Name:    timer.Name,
					Version: timer.Version,
					To:      timer.To,
					Input:   timer.Input,
				})
				if err != nil {
					log.WithField("error", err.Error()).Error("Unable to encode a sender event")
					continue
				}

				if err := prod.Publish("sender_"+timer.Sender, body); err != nil {
					log.WithField("error", err.Error()).Error("Unable to send an event")
					continue
				}

				// Delete it from RDB state
				if err := r.Db(*rethinkdbDatabase).Table("hub_state").Get(timer.ID).Delete().Exec(session); err != nil {
					log.WithField("error", err.Error()).Error("Unable to remove an event from database")
					continue
				}

				timersToDelete = append(timersToDelete, id)
			} else {
				break
			}
		}
		for y, x := range timersToDelete {
			i := x - y
			copy(state[i:], state[i+1:])
			state[len(state)-1] = nil
			state = state[:len(state)-1]
		}
		stateLock.Unlock()

		if len(state) > 0 {
			select {
			case <-time.After(state[0].Time.Sub(time.Now())):
				break
			case <-change:
				break
			}
		} else {
			<-change
		}
	}
}
