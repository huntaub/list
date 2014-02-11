package main

import (
	"fmt"
	"schedule"
	// "strconv"
	"bufio"
	"flag"
	"io"
	"net/http"
	"os"
	"time"
)

var cache *string = flag.String("cache", "", "Location of the cached Lou's List")
var classes *string = flag.String("classes", "", "Location of the cached Lou's List")

var classList schedule.ClassList

func main() {
	flag.Parse()
	fmt.Println("Welcome to the Cicus of Value!")
	fmt.Println("Time to enter some classes. Enter 'done' when finished. Use the following format:")
	fmt.Println("HIUS 2061")
	var err error

	// myClasses := schedule.CreateSchedule()
	schedulizer := schedule.CreateSchedulizer()

	var update func(time.Time) = func(now time.Time) {
		var theSchedule io.Reader

		fmt.Println(now, ": Updating Class List")
		if *cache == "" {
			// resp, err := http.Get("http://rabi.phys.virginia.edu/mySIS/CS2/page.php?Semester=1142&Type=Group&Group=CS&Print=")
			resp, err := http.Get("http://rabi.phys.virginia.edu/mySIS/CS2/page.php?Semester=1142&Type=Group&Group=Economics&Print=")
			if err != nil {
				panic(err)
			}
			fmt.Println("Got Lou's List")
			defer resp.Body.Close()
			theSchedule = resp.Body
		} else {
			fi, err := os.Open(*cache)
			if err != nil {
				panic(err)
			}
			// close fi on exit and check for its returned error
			defer func() {
				if err := fi.Close(); err != nil {
					panic(err)
				}
			}()
			theSchedule = fi
		}

		var err error
		classList, err = schedule.ParseList(theSchedule)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	if *cache == "" {
		go func() {
			c := time.Tick(1 * time.Hour)
			for now := range c {
				update(now)
			}
		}()
	}
	update(time.Now())

	var reader io.Reader = os.Stdin
	if *classes != "" {
		reader, err = os.Open(*classes)
		if err != nil {
			panic(err)
		}
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "done" {
			break
		}

		c, ok := classList[text]
		if !ok {
			fmt.Println("Couldn't find that class.")
			continue
		}
		fmt.Println(c.Sections[0].Enrollment)
		schedulizer.AddClass(c)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	for _, class := range schedulizer.Calculate() {
		fmt.Println("Found Schedule:", class)
	}
}
