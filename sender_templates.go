package main

import (
	htl "html/template"
	"sort"
	"sync"
	ttl "text/template"

	"github.com/Sirupsen/logrus"
	"github.com/blang/semver"
	r "github.com/dancannon/gorethink"
)

var (
	templateLock sync.RWMutex

	templates        map[string]map[string]*Template
	templateVersions map[string]semver.Versions
)

type Template struct {
	ID      string `gorethink:"id"`
	Name    string `gorethink:"name"`
	Version string `gorethink:"version"`
	Subject string `gorethink:"subject"`
	Body    string `gorethink:"body"`

	SubjectTpl *ttl.Template `gorethink:"-"`
	BodyTpl    *htl.Template `gorethink:"-"`
}

func parseTemplate(tpl *Template) {
	// Parse the version
	version, err := semver.Parse(tpl.Version)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":      tpl.ID,
			"version": tpl.Version,
			"error":   err.Error(),
		}).Error("Unable to parse template's version")
		return
	}

	// Parse the subject template
	stpl, err := ttl.New(tpl.Name + "_subject").Parse(tpl.Subject)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":    tpl.ID,
			"error": err.Error(),
		}).Error("Unable to parse template's subject")
		return
	}
	tpl.SubjectTpl = stpl

	// Parse the body template
	btpl, err := htl.New(tpl.Name + "_body").Parse(tpl.Body)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":    tpl.ID,
			"error": err.Error(),
		}).Error("Unable to parse templates's body")
		return
	}
	tpl.BodyTpl = btpl

	// Prepare template storages
	if _, ok := templates[tpl.Name]; !ok {
		templates[tpl.Name] = map[string]*Template{}
	}
	if _, ok := templateVersions[tpl.Name]; !ok {
		templateVersions[tpl.Name] = semver.Versions{}
	}

	// Put it into the templates storage
	templates[tpl.Name][version.String()] = tpl
	templateVersions[tpl.Name] = append(templateVersions[tpl.Name], version)

	log.Printf("Loaded template %s %s", tpl.Name, tpl.Version)
}

func loadTemplates() {
	cursor, err := r.Db(*rethinkdbDatabase).Table("templates").Run(session)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Enable to query for templates")
	}

	var tpls []*Template
	if err := cursor.All(&tpls); err != nil {
		log.WithField("error", err.Error()).Fatal("Enable to map templates into a slice")
	}

	for _, tpl := range tpls {
		parseTemplate(tpl)
	}
}

func deleteTemplate(tpl *Template) error {
	// Parse old template's version
	version, err := semver.Parse(tpl.Version)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":      tpl.ID,
			"version": tpl.Version,
			"error":   err.Error(),
		}).Error("Unable to parse template's version")
		return err
	}

	// Remove the old template
	delete(templates[tpl.Name], version.String())
	for i, v2 := range templateVersions[tpl.Name] {
		if v2.Equals(version) {
			templateVersions[tpl.Name] = append(templateVersions[tpl.Name][:i], templateVersions[tpl.Name][i+1:]...)
			break
		}
	}

	return nil
}

func sortVersions(what string) {
	if what == "" {
		for _, versions := range templateVersions {
			sort.Sort(versions)
		}
	} else {
		if versions, ok := templateVersions[what]; ok {
			sort.Sort(versions)
		}
	}
}

func initTemplates() {
	templateLock.Lock()

	templates = map[string]map[string]*Template{}
	templateVersions = map[string]semver.Versions{}

	loadTemplates()
	sortVersions("")

	templateLock.Unlock()

	cursor, err := r.Db(*rethinkdbDatabase).Table("templates").Changes().Run(session)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Enable to query for template changes")
	}

	var change struct {
		NewValue *Template `gorethink:"new_val"`
		OldValue *Template `gorethink:"old_val"`
	}
	for cursor.Next(&change) {
		templateLock.Lock()
		if change.OldValue != nil && change.NewValue == nil {
			if err := deleteTemplate(change.OldValue); err != nil {
				continue
			}
		} else if change.OldValue == nil && change.NewValue != nil {
			parseTemplate(change.NewValue)
			sortVersions(change.NewValue.Name)
		} else {
			if err := deleteTemplate(change.OldValue); err != nil {
				continue
			}
			parseTemplate(change.NewValue)
			sortVersions(change.NewValue.Name)
		}

		templateLock.Unlock()
	}
}
