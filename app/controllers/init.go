package controllers

import (
	"github.com/huntaub/list/schedule"
	"github.com/robfig/revel"
	"labix.org/v2/mgo"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
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
	StartParser()

	// Regex to Recognize Classes
	revel.INFO.Printf("Compiling Regular Expressions...")
	classRegex = regexp.MustCompile(`([A-z]{1,4})\s?(\d{4})\s?(?::{((?:,?\s?\d{1,3})+)})?`)
	sectionRegex = regexp.MustCompile(`\d{1,3}`)

	revel.INFO.Printf("Adding Template Functions...")
	CreateTemplateFunctions()

	// Interceptions
	revel.INFO.Printf("Starting Interceptors...")
	revel.InterceptMethod(App.Init, revel.BEFORE)

	revel.INFO.Printf("Initialization Complete")
}

// Start Parser
func StartParser() {
	f := func(now time.Time) {
		revel.INFO.Printf("Updating Lou's List at %v", now)
		resp, err := http.Get("http://rabi.phys.virginia.edu/mySIS/CS2/page.php?Semester=1148&Type=Group&Group=CS&Print=")
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
	go func() {
		c := time.Tick(time.Hour * 1)
		for {
			now := <-c
			f(now)
		}
	}()
	f(time.Now())
}

// Prepopulate Revel with Important Template Functions
func CreateTemplateFunctions() {
	// Get the Last Name of an Instructor
	revel.TemplateFuncs["lastName"] = func(a string) string {
		if a == "Staff" {
			return a
		}
		nameComp := strings.Split(a, " ")
		if len(nameComp) != 2 {
			return a
		} else {
			return nameComp[1]
		}
	}

	// Correctly Format the Time of a Class
	revel.TemplateFuncs["formatTime"] = func(a time.Time) string {
		str := a.Format(time.RFC1123)
		return str[:len(str)-3] + "EST"
	}

	// Adds One to a Variable
	revel.TemplateFuncs["addOne"] = func(a int) int {
		return a + 1
	}

	// Returns the Last time the Page was Updated
	revel.TemplateFuncs["lastUpdated"] = func() string {
		return lastUpdate.Format("January 2, 3:04PM")
	}

	// Creates a Context to Send to views/class.html
	revel.TemplateFuncs["classCreator"] = func(class interface{}, loggedIn interface{}) map[string]interface{} {
		return map[string]interface{}{
			"class":    class,
			"loggedIn": loggedIn,
		}
	}

	// Apply Border Format Based on Class Capacity
	revel.TemplateFuncs["sectionBorder"] = func(v *schedule.Section) string {
		totalEnrollment := float64(v.Enrollment)
		totalCapacity := float64(v.Capacity)
		if totalEnrollment/totalCapacity >= 1 {
			return "border-danger"
		} else if totalEnrollment/totalCapacity >= 0.5 {
			return "border-warning"
		}
		return "border-default"
	}
}
