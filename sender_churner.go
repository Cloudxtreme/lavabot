package main

type SenderEvent struct {
	Name    string      `gorethink:"name"`
	Version string      `gorethink:"version"`
	To      []string    `gorethink:"to"`
	Input   interface{} `gorethink:"input"`
}

func initSender() {
	// ayy lmao
}
