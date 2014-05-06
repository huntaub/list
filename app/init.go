package app

import (
	"github.com/huntaub/list/app/controllers"
	"github.com/huntaub/list/app/schedule"
	"github.com/robfig/revel"
	"strings"
	"time"
)

func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.ActionInvoker,           // Invoke the action.
	}

	// Begin Application Startup
	revel.OnAppStart(controllers.StartApp)

	CreateTemplateFunctions()
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
