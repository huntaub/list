package controllers

import (
	"github.com/huntaub/list/schedule"
	/*	"github.com/robfig/revel"*/
	"labix.org/v2/mgo"
)

func SaveClassesToDB(c schedule.ClassList, collection *mgo.Collection) {
	/*	if revel.Config.BoolDefault("list.db.overwrite", false) {
		collection.RemoveAll(nil)
		for _, v := range c {
			collection.Insert(v)
		}
	}*/
}
