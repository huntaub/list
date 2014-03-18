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
	now := time.Now()
	revel.INFO.Printf("%s - Searching %s", now.String(), class)

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

func (c App) SchedulesFromList(list string) ([]*schedule.Schedule, revel.Result) {
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

	return schedulizer.Calculate(), nil
}

func (c App) Build(userList string) revel.Result {
	now := time.Now()
	revel.INFO.Printf("%s - Building %s", now.String(), userList)

	sched, r := c.SchedulesFromList(userList)
	if r != nil {
		return r
	}

	if len(sched) > 20 {
		sched = sched[:20]
	}

	c.RenderArgs = map[string]interface{}{
		"sched": sched,
		"perma": base64.URLEncoding.EncodeToString([]byte(userList)),
	}

	return c.Render()
}

func (c App) Schedule(perm string, num int) revel.Result {
	blah, _ := base64.URLEncoding.DecodeString(perm)

	now := time.Now()
	revel.INFO.Printf("%s - Fetching %s : %d", now.String(), blah, num)

	sched, r := c.SchedulesFromList(string(blah))
	if r != nil {
		return r
	}

	c.RenderArgs = map[string]interface{}{
		"sched": []*schedule.Schedule{sched[num]},
		"perma": perm,
	}

	return c.RenderTemplate("App/Build.html")
}
