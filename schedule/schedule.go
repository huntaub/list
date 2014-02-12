package schedule

import (
	"fmt"
	"strconv"
	"time"
)

type Schedulizer struct {
	ListedClasses map[string][]ClassTime
}

func CreateSchedulizer() *Schedulizer {
	return &Schedulizer{
		ListedClasses: make(map[string][]ClassTime),
	}
}

func (s *Schedulizer) AddClass(c *Class) {
	s.ListedClasses[c.Department+" "+strconv.Itoa(c.Number)] = c.ValidClassTimes()
}

func (s *Schedulizer) Calculate() []*Schedule {
	// Make a New Schedule for Each Ambigious Time Slot
	// Delete Schedules that Don't Allow any new Class Times
	// fmt.Println("Finding Schedule...")
	allSchedules := []*Schedule{}
	for _, v := range s.ListedClasses {
		allSchedules = CreateSchedulesByAddingClass(allSchedules, v)
		if len(allSchedules) == 0 {
			// fmt.Println("No possible Schedule")
			return nil
		}
	}
	return allSchedules
}

func CreateSchedulesByAddingClass(old []*Schedule, add []ClassTime) []*Schedule {
	output := []*Schedule{}
	if len(old) == 0 {
		old = []*Schedule{new(Schedule)}
	}
	for _, s := range old {
		for _, v := range add {
			sched := new(Schedule)
			sched.Classes = make([]ClassTime, len(s.Classes))
			copy(sched.Classes, s.Classes)
			if sched.AddClass(v) {
				output = append(output, sched)
			}
		}
	}
	return output
}

type ClassTime struct {
	*Class
	SectionNumber []int
}

func (c ClassTime) GetSections() []*Section {
	output := []*Section{}
	c.CreateSectionMap()
	for _, v := range c.SectionNumber {
		output = append(output, c.SectionMap[v])
	}
	return output
}

func (c ClassTime) ConflictsWithClassTime(b ClassTime) bool {
	c.Class.CreateSectionMap()
	b.Class.CreateSectionMap()
	for _, sect1 := range c.SectionNumber {
		for _, sect2 := range b.SectionNumber {
			// fmt.Println(sect1, sect2)
			if c.SectionMap[sect1].Overlaps(b.SectionMap[sect2]) {
				return true
			}
		}
	}
	return false
}

func (c ClassTime) String() string {
	output := "<" + c.Class.Department + " " + strconv.Itoa(c.Class.Number) + ": {"
	for i, x := range c.SectionNumber {
		if i != 0 {
			output += ", "
		}
		output += strconv.Itoa(x)
	}
	return output + "}>"
}

func (c ClassTime) PrintForCalendar(line int, meetingIndex int) string {
	return c.Class.PrintForCalendar(line, c.SectionNumber[0], meetingIndex)
}

func (c ClassTime) Equals(x ClassTime) bool {
	return c.Class.Equals(x.Class)
}

type RankingAlgorithm func(c []ClassTime) int

type Schedule struct {
	Classes []ClassTime
}

func CreateSchedule() *Schedule {
	return &Schedule{
		Classes: make([]ClassTime, 0),
	}
}

func (s *Schedule) ScoreSchedule(r RankingAlgorithm) int {
	return r(s.Classes)
}

func (s *Schedule) AddClass(c ClassTime) bool {
	for _, v := range s.Classes {
		if v.ConflictsWithClassTime(c) {
			return false
		}
	}
	s.Classes = append(s.Classes, c)
	return true
}

func (s *Schedule) ContainsClass(c ClassTime) bool {
	for _, v := range s.Classes {
		if v.Class.Equals(c.Class) {
			return true
		}
	}
	return false
}

func (s *Schedule) NumClasses() int {
	return len(s.Classes)
}

func (s *Schedule) String() string {
	output := "[ "
	for _, v := range s.Classes {
		output += v.String() + " "
	}
	return output + "]"
}

func (s *Schedule) Print() {
	printTime, _ := time.Parse("15:04", "7:50")

	var lastClass []ClassTime = make([]ClassTime, 6)
	var lineNumber []int = make([]int, 6)

	for h := 0; h < 74; h += 1 {
		fmt.Print("|")
		for i := 0; i <= 5; i += 1 {

			length := 15
			if i == 0 {
				length = 7
			}

			if h == 0 || h == 73 {
				for j := 0; j < length; j += 1 {
					fmt.Print("-")
				}
			} else {
				if i == 0 {
					fmt.Print(printTime.Format(" 15:04 "))
				} else {
					c, index, ok := s.ClassAtTime(printTime, time.Weekday(i))

					if !c.Equals(lastClass[i]) {
						lastClass[i] = c
						lineNumber[i] = -1
					}

					if !ok {
						for j := 0; j < length; j += 1 {
							fmt.Print(" ")
						}
					} else {
						if c.Equals(lastClass[i]) {
							lineNumber[i] += 1
						}
						output := "   " + c.PrintForCalendar(lineNumber[i], index)
						// Pad the Output
						for len(output) < length {
							output += " "
						}
						fmt.Print(output)
					}
				}
			}
			fmt.Print("|")
		}
		fmt.Print("\n")
		printTime = printTime.Add(10 * time.Minute)
	}
}

func (s *Schedule) ClassAtTime(t time.Time, day time.Weekday) (ClassTime, int, bool) {
	for _, v := range s.Classes {
		theSection := v.Sections[v.SectionNumber[0]]
		for i, x := range theSection.Meetings {
			if TimeBeforeOrEqualTo(x.StartTime, t) && TimeAfterOrEqualTo(x.EndTime, t) && x.Days.Contains(day) {
				return v, i, true
			}
		}
	}
	return ClassTime{}, 0, false
}

func TimeBeforeOrEqualTo(x time.Time, y time.Time) bool {
	return x.Before(y) || x == y
}

func TimeAfterOrEqualTo(x time.Time, y time.Time) bool {
	return x.After(y) || x == y
}
