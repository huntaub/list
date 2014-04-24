package schedule

import "labix.org/v2/mgo"

func (c ClassList) SaveToDB(collection *mgo.Collection) {
	// collection.RemoveAll(nil)
	// for _, v := range c {
	// 	collection.Insert(v)
	// }
}
