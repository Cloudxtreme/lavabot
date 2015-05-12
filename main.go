package main

import (
	"strings"

	"github.com/Sirupsen/logrus"
	r "github.com/dancannon/gorethink"
	"github.com/namsral/flag"
)

var (
	configFlag       = flag.String("config", "", "Config file to read")
	logFormatterType = flag.String("log_formatter", "text", "Log formatter type")
	logForceColors   = flag.Bool("log_force_colors", false, "Force colored prompt?")

	apiURL = flag.String("api_url", "https://api.lavaboom.com", "Path to the Lavaboom API")

	nsqdAddress       = flag.String("nsqd_address", "127.0.0.1:4150", "Address of the NSQ server")
	lookupdAddress    = flag.String("lookupd_address", "127.0.0.1:4160", "Address of the nsqlookupd server")
	rethinkdbAddress  = flag.String("rethinkdb_address", "127.0.0.1:28015", "Address of the RethinkDB server")
	rethinkdbDatabase = flag.String("rethinkdb_database", "lavabot", "RethinkDB database to use")

	enableHub      = flag.Bool("enable_hub", true, "Enable hub module")
	enableReceiver = flag.Bool("enable_receiver", true, "Enable receiver module")
	enableSender   = flag.Bool("enable_sender", true, "Enable sender module")

	welcomeName           = flag.String("welcome_name", "welcome", "Name of the welcome template to use")
	welcomeVersion        = flag.String("welcome_version", "1.0.0", "Version of the welcome template to use")
	gettingStartedName    = flag.String("getting_name", "getting", "Name of the getting started template to use")
	gettingStartedVersion = flag.String("getting_version", "1.0.0", "Version of the getting started template to use")
	securityName          = flag.String("security_name", "security", "Name of the security info template to use")
	securityVersion       = flag.String("security_version", "1.0.0", "Version of the security info template to use")
	whatsUpName           = flag.String("whatsup_name", "whatsup", "Name of the how's it going template to use")
	whatsUpVersion        = flag.String("whatsup_version", "1.0.0", "Version of the how's it going template to use")

	usernames        = flag.String("usernames", "", "Usernames to use in the sender")
	passwords        = flag.String("passwords", "", "Passwords to use in the sender")
	privateKeys      = flag.String("private_keys", "", "Private keys to use in the receiver")
	grooveAddress    = flag.String("groove_address", "", "Address of the Groove forwarding email")
	forwardingServer = flag.String("forwarding_server", "127.0.0.1:25", "Address of the SMTP server used for email forwarding")
	dkimKey          = flag.String("dkim_key", "./dkim.key", "Path of the DKIM key")
)

var (
	session *r.Session
	log     *logrus.Logger
)

func main() {
	flag.Parse()

	log = logrus.New()
	if *logFormatterType == "text" {
		log.Formatter = &logrus.TextFormatter{
			ForceColors: *logForceColors,
		}
	} else if *logFormatterType == "json" {
		log.Formatter = &logrus.JSONFormatter{}
	}
	log.Level = logrus.DebugLevel

	if *enableHub || *enableSender {
		var err error
		session, err = r.Connect(r.ConnectOpts{
			Address: *rethinkdbAddress,
		})
		if err != nil {
			log.WithField("error", err.Error()).Fatal("Unable to connect to RethinkDB")
		}

		r.DbCreate(*rethinkdbDatabase).Exec(session)
		r.Db(*rethinkdbDatabase).TableCreate("templates").Exec(session)
		r.Db(*rethinkdbDatabase).Table("templates").IndexCreate("name").Exec(session)
		r.Db(*rethinkdbDatabase).Table("templates").IndexCreate("version").Exec(session)
		r.Db(*rethinkdbDatabase).TableCreate("hub_state").Exec(session)
		r.Db(*rethinkdbDatabase).Table("hub_state").IndexCreate("time").Exec(session)
	}

	/*if *enableSender || *enableReceiver {
		var err error
		api, err = client.New(*apiURL, 0)
		if err != nil {
			log.WithField("error", err.Error).Fatal("Unable to connect to the Lavaboom API")
		}
	}*/

	up := strings.Split(*usernames, ",")
	pp := strings.Split(*passwords, ",")
	kp := strings.Split(*privateKeys, ",")

	if len(up) != len(pp) {
		log.Fatal("length of usernames and passwords is different")
	}

	if *enableReceiver && len(up) != len(kp) {
		log.Fatal("length of keys doesn't match the length of usernames")
	}

	if *enableReceiver {
		for i, username := range up {
			go initReceiver(username, pp[i], kp[i])
		}
	}

	if *enableSender {
		go initTemplates()

		for i, username := range up {
			go initSender(username, pp[i])
		}
	}

	if *enableHub {
		change := make(chan struct{})

		go initChurner(change)
		go initHub(change)
	}

	select {}
}
