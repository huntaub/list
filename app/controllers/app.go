package controllers

import (
	"encoding/base64"
	"fmt"
	"github.com/huntaub/list/app/routes"
	"github.com/huntaub/list/schedule"
	"github.com/robfig/revel"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var classList schedule.ClassList
var lastUpdate time.Time
var classRegex *regexp.Regexp

func init() {
	f := func(now time.Time) {
		fmt.Println("Updating Lou's List at", now)
		// fi, err := os.Open("/Users/hunter/Documents/Developer/golang/src/list/app/cache/complete_schedule.html")
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
	classRegex = regexp.MustCompile(`([A-z]{2,4})\s?(\d+)`)

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

	revel.TemplateFuncs["formatTime"] = func(a time.Time) string {
		str := a.Format(time.RFC1123)
		return str[:len(str)-3] + "EST"
	}

	revel.TemplateFuncs["addOne"] = func(a int) int {
		return a + 1
	}
}

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) NotFound() revel.Result {
	return c.Render()
}

func (c App) Class(dept string, num string) revel.Result {
	cl, ok := classList[dept+" "+num]

	if !ok {
		return c.Redirect(routes.App.NotFound())
	}

	c.RenderArgs = map[string]interface{}{
		"class":       cl,
		"ok":          ok,
		"lastUpdated": lastUpdate,
	}

	return c.Render()
}

func (c App) Search(class string) revel.Result {
	revel.INFO.Printf("Searching %s", class)

	matches := classRegex.FindAllStringSubmatch(class, -1)
	if len(matches) < 1 || len(matches[0]) < 3 {
		return c.Redirect(routes.App.NotFound())
	}
	lookup := strings.ToUpper(matches[0][1]) + " " + matches[0][2]

	_, ok := classList[lookup]

	if !ok {
		return c.Redirect(routes.App.NotFound())
	} else {
		comp := strings.Split(lookup, " ")
		return c.Redirect(routes.App.Class(comp[0], comp[1]))
	}
}

func (c App) SchedulesFromList(list string, stop chan bool) ([]*schedule.Schedule, revel.Result) {
	matches := classRegex.FindAllStringSubmatch(list, -1)
	if len(matches) < 0 {
		return nil, c.Redirect(routes.App.NotFound())
	} else if len(matches) == 1 {
		return nil, c.Redirect(routes.App.Class(strings.ToUpper(matches[0][1]), matches[0][2]))
	}
	schedulizer := schedule.CreateSchedulizer()

	for _, class := range matches {
		if len(class) != 3 {
			return nil, c.Redirect(routes.App.NotFound())
		}

		lookup := strings.ToUpper(class[1]) + " " + class[2]
		cl, ok := classList[lookup]
		if !ok {
			c.Flash.Error("Couldn't find " + lookup + " in UVa class listings.")
			return nil, c.Redirect(routes.App.NotFound())
		}

		schedulizer.AddClass(cl)
	}
	out := schedulizer.Calculate(stop)

	return out, nil
}

func (c App) Build(userList string) revel.Result {
	revel.INFO.Printf("Building %s", userList)

	var (
		sched []*schedule.Schedule
		r     revel.Result
	)

	timeout := time.After(3 * time.Second)
	timedout := false
	done := make(chan bool, 1)
	quit := make(chan bool, 1)

	go func() {
		sched, r = c.SchedulesFromList(userList, quit)
		revel.INFO.Printf("Successfully finished building.")
		done <- true
	}()

	select {
	case <-timeout:
		quit <- true
		time.Sleep(1 * time.Second)
		revel.INFO.Printf("Timed Out.")
		timedout = true
		// c.Flash.Error("Schedules may not be the best as computation timed out.")
	case <-done:
		// Just waiting on the world to change.
	}

	if r != nil {
		return r
	}

	tot := len(sched)
	if len(sched) > 20 {
		sched = sched[:20]
	}

	c.RenderArgs = map[string]interface{}{
		"to":    timedout,
		"total": tot,
		"sched": sched,
		"perma": base64.URLEncoding.EncodeToString([]byte(userList)),
	}

	return c.Render()
}

func (c App) Schedule(perm string, num int) revel.Result {
	blah, _ := base64.URLEncoding.DecodeString(perm)

	revel.INFO.Printf("Fetching %s : %d", blah, num)

	sched, r := c.SchedulesFromList(string(blah), make(chan bool, 1))
	if r != nil {
		return r
	}

	c.RenderArgs = map[string]interface{}{
		"sched": []*schedule.Schedule{sched[num]},
		"perma": perm,
	}

	return c.RenderTemplate("App/Build.html")
}

func (c App) Timeout() revel.Result {
	return c.Render()
}
