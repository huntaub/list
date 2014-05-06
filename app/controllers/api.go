package controllers

import (
	"github.com/huntaub/list/app/models"
	"github.com/robfig/revel"
)

type API struct {
	App
}

func (c API) Documentation() revel.Result {
	return c.Render()
}

func (c API) Settings() revel.Result {
	var u models.User
	err := users.Find(map[string]string{"email": c.Session["user"]}).One(&u)
	if err != nil {
		panic(err)
	}
	c.RenderArgs["user"] = u
	return c.Render()
}

func (c API) Class() revel.Result      { return c.Render() }
func (c API) Department() revel.Result { return c.Render() }
func (c API) Schedule() revel.Result   { return c.Render() }
