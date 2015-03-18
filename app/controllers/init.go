package controllers

import (
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/huntaub/list/app/schedule"
	"github.com/revel/revel"
	"github.com/revel/revel/modules/jobs/app/jobs"
	"labix.org/v2/mgo"
)

var classList schedule.ClassList
var lastUpdate time.Time
var classRegex, sectionRegex *regexp.Regexp

var collection *mgo.Collection
var lists *mgo.Collection
var users *mgo.Collection

func StartApp() {
	revel.INFO.Printf("Beginning Initialization Process...")

	// Load Database Connections
	revel.INFO.Printf("Connecting to Database...")
	session, err := mgo.Dial("mongodb://leath:hunter0813@oceanic.mongohq.com:10000/list")
	if err != nil {
		panic(err)
	}

	collection = session.DB("list").C("classes")
	users = session.DB("list").C("users")
	rand.Seed(time.Now().UnixNano())

	// Start Parsing Lou's List
	revel.INFO.Printf("Launching Parser...")
	jobs.Now(Parser{})
	jobs.Every(1*time.Hour, Parser{})

	// Regex to Recognize Classes
	revel.INFO.Printf("Compiling Regular Expressions...")
	classRegex = regexp.MustCompile(`([A-z]{1,4})\s?(\d{4})\s?(?::{((?:,?\s?\d{1,3})+)})?`)
	sectionRegex = regexp.MustCompile(`\d{1,3}`)

	/*	revel.INFO.Printf("Adding Template Functions...")
		CreateTemplateFunctions()*/

	// Interceptions
	revel.INFO.Printf("Starting Interceptors...")
	revel.InterceptMethod(App.Init, revel.BEFORE)

	revel.INFO.Printf("Initialization Complete")
}

type Parser struct{}

// Start Parser
func (l Parser) Run() {
	now := time.Now()
	revel.INFO.Printf("Updating Lou's List at %v", now)
	resp, err := http.Get("http://rabi.phys.virginia.edu/mySIS/CS2/page.php?Semester=1158&Type=Group&Group=CS&Print=")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	classList, err = schedule.ParseList(resp.Body)
	if err != nil {
		panic(err)
	}
	SaveClassesToDB(classList, collection)
	lastUpdate = now
}
