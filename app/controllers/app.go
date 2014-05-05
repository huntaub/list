package controllers

import (
	"encoding/base64"
	"fmt"
	"github.com/huntaub/list/app/routes"
	"github.com/huntaub/list/schedule"
	"github.com/robfig/revel"
	"sort"
	"strconv"
	"strings"
	"time"
)

type App struct {
	*revel.Controller
	session string
}

// Automatically Load Login State from Session
func (c App) Init() revel.Result {
	_, loggedIn := c.Session["user"]
	c.RenderArgs["loggedIn"] = loggedIn
	return nil
}

// Index of the Site
func (c App) Index() revel.Result {
	var result []string
	err := collection.Find(nil).Distinct("department", &result)
	if err != nil {
		panic(err)
	}

	user, ok := c.Session["user"]
	var classes []*schedule.Class
	if ok {
		var u User
		err := users.Find(map[string]string{"email": user}).One(&u)
		if err != nil {
			panic(err)
		}
		classes = make([]*schedule.Class, len(u.ClassBucket))
		for i, v := range u.ClassBucket {
			classes[i] = classList[v]
		}
	}

	sort.Sort(sort.StringSlice(result))

	return c.Render(result, classes)
}

func (c App) NotFound() revel.Result {
	return c.Render()
}

// "sections": {"$elemMatch": {"instructor": "Aaron Bloomfield"}}

//"sections":{"$elemMatch":{"sisnumber":10120}}

func (c App) Class(dept string, num int) revel.Result {
	var class *schedule.Class
	err := collection.Find(map[string]interface{}{"department": dept, "number": num}).One(&class)
	if err != nil {
		fmt.Println(err)
		return c.Redirect(routes.App.NotFound())
	}

	c.RenderArgs["class"] = class

	return c.Render()
}

func (c App) Department(dept string) revel.Result {
	var classes []schedule.Class
	err := collection.Find(map[string]string{"department": dept}).Sort("number").All(&classes)
	if err != nil {
		panic(err)
	}

	return c.Render(classes, dept)
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
		number, _ := strconv.Atoi(comp[1])
		return c.Redirect(routes.App.Class(comp[0], number))
	}
}

func (c App) SchedulesFromList(list string, stop chan bool) ([]*schedule.Schedule, revel.Result) {
	matches := classRegex.FindAllStringSubmatch(list, -1)
	if len(matches) < 0 {
		return nil, c.Redirect(routes.App.NotFound())
	} else if len(matches) == 1 {
		number, _ := strconv.Atoi(matches[0][2])
		return nil, c.Redirect(routes.App.Class(strings.ToUpper(matches[0][1]), number))
	}
	schedulizer := schedule.CreateSchedulizer()

	for _, class := range matches {
		if len(class) < 3 {
			return nil, c.Redirect(routes.App.NotFound())
		}

		lookup := strings.ToUpper(class[1]) + " " + class[2]
		cl, ok := classList[lookup]

		if class[3] != "" {
			sections := sectionRegex.FindAllString(class[3], -1)
			// revel.INFO.Printf(class[3])

			var newClass *schedule.Class = new(schedule.Class)
			*newClass = *cl

			newClass.SectionMap = nil
			newClass.Sections = nil
			cl.CreateSectionMap()
			// revel.INFO.Printf(cl.Sections.String())

			for _, v := range sections {
				i, err := strconv.Atoi(v)
				revel.INFO.Printf("%v %v", v, i)
				if err != nil {
					revel.WARN.Printf(err.Error())
					c.Flash.Error(lookup + " section " + v + " isn't a real section.")
					return nil, c.Redirect(routes.App.NotFound())
				}

				s, ok := cl.SectionMap[i]
				if !ok {
					c.Flash.Error("Couldn't find " + lookup + " section " + v + " in UVa class listings.")
					return nil, c.Redirect(routes.App.NotFound())
				}

				newClass.Sections = append(newClass.Sections, s)
			}
			newClass.CreateSectionMap()

			revel.INFO.Printf(fmt.Sprintf("%v", newClass.ValidClassTimes()))

			cl = newClass
		}
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
		"to":      timedout,
		"total":   tot,
		"sched":   sched,
		"perma":   base64.URLEncoding.EncodeToString([]byte(userList)),
		"request": userList,
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

// Add Class to User Class List
func (c App) ClassAdd(dept string, num int) revel.Result {
	user, ok := c.Session["user"]
	if !ok {
		c.Response.Status = 403
		return c.Render()
	}

	var u User
	err := users.Find(map[string]string{"email": user}).One(&u)
	if err != nil {
		panic(err)
	}

	u.ClassBucket = append(u.ClassBucket, fmt.Sprintf("%v %v", dept, num))

	err = users.Update(map[string]string{"email": user}, u)
	if err != nil {
		panic(err)
	}

	return c.Redirect(routes.App.Index())
}

// Remove Class from User Class List
func (c App) ClassRemove(dept string, num int) revel.Result {
	user, ok := c.Session["user"]
	if !ok {
		c.Response.Status = 403
		return c.Render()
	}

	var u User
	err := users.Find(map[string]string{"email": user}).One(&u)
	if err != nil {
		panic(err)
	}

	newBucket := make([]string, len(u.ClassBucket)-1)
	found := false
	i := 0
	for _, v := range u.ClassBucket {
		if v == fmt.Sprintf("%v %v", dept, num) && !found {
			found = true
			continue
		}
		newBucket[i] = v
		i++
	}
	u.ClassBucket = newBucket

	err = users.Update(map[string]string{"email": user}, u)
	if err != nil {
		panic(err)
	}

	return c.Redirect(routes.App.Index())
}
