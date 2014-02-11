package controllers

import (
	"fmt"
	"github.com/robfig/revel"
	"list/app/routes"
	"list/schedule"
	"os"
	"strings"
	"time"
)

var classList schedule.ClassList
var lastUpdate time.Time

func init() {
	f := func(now time.Time) {
		fmt.Println("Updating Lou's List at", now)
		fi, err := os.Open("/Users/hunter/Documents/Developer/golang/src/list/app/cache/complete_schedule.html")
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := fi.Close(); err != nil {
				panic(err)
			}
		}()

		classList, err = schedule.ParseList(fi)
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
	_, ok := classList[class]

	if !ok {
		return c.Redirect(routes.App.NotFound())
	} else {
		comp := strings.Split(class, " ")
		return c.Redirect(routes.App.Class(comp[0], comp[1]))
	}
}
