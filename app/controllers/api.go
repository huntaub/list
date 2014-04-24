package controllers

import (
	"github.com/robfig/revel"
)

type API struct {
	App
}

func (c API) Documentation() revel.Result {
	return c.Render()
}
