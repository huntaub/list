package controllers

import (
	"fmt"
	"github.com/huntaub/list/app/schedule"
	"github.com/revel/revel"
	"labix.org/v2/mgo"
)

type DBRecord struct {
	Department string
	Number     int
}

func SaveClassesToDB(c schedule.ClassList, collection *mgo.Collection) {
	if revel.Config.BoolDefault("server.dbwrite", false) {
		revel.INFO.Printf("Writing Data to Database")

		// Get all Current Records
		var records []DBRecord
		err := collection.Find(nil).Select(map[string]int{"department": 1, "number": 1}).All(&records)
		if err != nil {
			revel.ERROR.Printf("Unable to get records from database: %v", err)
			return
		}

		// Put them in a Hash Map
		recordsMap := make(map[string]DBRecord)
		for _, v := range records {
			recordsMap[fmt.Sprintf("%v %v", v.Department, v.Number)] = v
		}

		// Iterate over new records
		for _, v := range c {
			// Not Guaranteed to Work
			collection.Upsert(map[string]interface{}{"department": v.Department, "number": v.Number}, v)

			delete(recordsMap, fmt.Sprintf("%v %v", v.Department, v.Number))
		}

		// Remove Classes that are Gone
		for _, v := range recordsMap {
			collection.Remove(map[string]interface{}{"department": v.Department, "number": v.Number})
		}

	}
}
