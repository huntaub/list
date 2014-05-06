package controllers

import (
	"encoding/json"
	"github.com/huntaub/list/app/models"
	"github.com/huntaub/list/app/schedule"
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

func (c API) getAPIObject() (map[string]interface{}, error) {
	var apiObject map[string]interface{}
	var outError error

	for key, _ := range c.Params.Form {
		outError = json.Unmarshal([]byte(key), &apiObject)
		if outError == nil {
			break
		}
	}

	return apiObject, outError
}

// API Class Object
func (c API) Class(key string) revel.Result {
	// Get API Object from Params
	apiObject, err := c.getAPIObject()
	if err != nil {
		return c.RenderJson(map[string]string{"error": "Unable to decode JSON request."})
	}

	var (
		tmp interface{}
		ok  bool
	)

	revel.INFO.Printf("%v", apiObject)

	// Verifications
	tmp, ok = apiObject["department"]
	if !ok {
		return c.RenderJson(map[string]string{"error": "Department field is required for Class method."})
	}

	department, ok := tmp.(string)
	if !ok {
		return c.RenderJson(map[string]string{"error": "Department field must be string."})
	}

	tmp, ok = apiObject["number"]
	if !ok {
		return c.RenderJson(map[string]string{"error": "Number field is required for Class method."})
	}

	number, ok := tmp.(float64)
	if !ok {
		return c.RenderJson(map[string]string{"error": "Number field must be integer."})
	}

	// Lookup class in Database
	var classObject *schedule.Class
	err = collection.Find(map[string]interface{}{"department": department, "number": int(number)}).One(&classObject)
	if err != nil {
		revel.ERROR.Printf("%v", err)
		return c.RenderJson(map[string]string{"error": "Couldn't find that department/class combination."})
	}

	return c.RenderJson(classObject)
}
func (c API) Department(key string) revel.Result { return c.Render() }
func (c API) Schedule(key string) revel.Result   { return c.Render() }
