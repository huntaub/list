package schedule

import (
	// "fmt"
	"sort"
	"strconv"
	"time"
)

type Schedulizer struct {
	ListedClasses map[string][]ClassTime
}

type ByRankingAlgorithm struct {
	Possible  []*Schedule
	Algorithm RankingAlgorithm
}

func (a ByRankingAlgorithm) Len() int { return len(a.Possible) }
func (a ByRankingAlgorithm) Swap(i, j int) {
	a.Possible[i], a.Possible[j] = a.Possible[j], a.Possible[i]
}
func (a ByRankingAlgorithm) Less(i, j int) bool {
	return a.Possible[i].ScoreSchedule(a.Algorithm) > a.Possible[j].ScoreSchedule(a.Algorithm)
}

func CreateSchedulizer() *Schedulizer {
	return &Schedulizer{
		ListedClasses: make(map[string][]ClassTime),
	}
}

var bestRanking RankingAlgorithm = func(c *Schedule) float64 {
	// Get a point for every class that isn't 8-12
	earlyPoints := 0.0
	for day := 1; day < 6; day++ {
		earlyClasses, _ := time.Parse("3:04PM", "8:00AM")
		for i := 0; i < 16; i++ {
			_, _, hasClass := c.ClassAtTime(earlyClasses, time.Weekday(day))
			if !hasClass {
				earlyPoints += 1
			} else {
				break
			}
			earlyClasses = earlyClasses.Add(time.Minute * 15)
		}
	}
	// Another point for every class that isn't 11:30-2 (30 minute increments)
	lunchPoints := 0.0
	for day := 1; day < 6; day++ {
		lunchClasses, _ := time.Parse("3:04PM", "11:30AM")
		for i := 0; i < 10; i++ {
			_, _, hasStartClass := c.ClassAtTime(lunchClasses.Add(time.Minute), time.Weekday(day))
			_, _, hasEndClass := c.ClassAtTime(lunchClasses.Add(time.Minute*29), time.Weekday(day))
			if !hasStartClass && !hasEndClass {
				lunchPoints += 1
			}
			lunchClasses = lunchClasses.Add(time.Minute * 15)
		}
	}
	// Another point for every class that isn't 3-10
	latePoints := 0.0
	for day := 1; day < 6; day++ {
		lateClasses, _ := time.Parse("3:04PM", "10:00PM")
		for i := 0; i < 28; i++ {
			_, _, hasClass := c.ClassAtTime(lateClasses, time.Weekday(day))
			if !hasClass {
				latePoints += 1
			} else {
				break
			}
			lateClasses = lateClasses.Add(time.Minute * -15)
		}
	}
	return (latePoints + earlyPoints + lunchPoints)
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
	sort.Sort(ByRankingAlgorithm{allSchedules, bestRanking})
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

type RankingAlgorithm func(c *Schedule) float64

type Schedule struct {
	Classes []ClassTime
	Score   float64
}

func CreateSchedule() *Schedule {
	return &Schedule{
		Classes: make([]ClassTime, 0),
	}
}

func (s *Schedule) ScoreSchedule(r RankingAlgorithm) float64 {
	s.Score = r(s)
	return s.Score
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

func (s *Schedule) ClassAtTime(t time.Time, day time.Weekday) (ClassTime, int, bool) {
	for _, v := range s.Classes {
		for _, m := range v.SectionNumber {
			for i, x := range v.SectionMap[m].Meetings {
				if TimeBeforeOrEqualTo(x.StartTime, t) && TimeAfterOrEqualTo(x.EndTime, t) && x.Days.Contains(day) {
					return v, i, true
				}
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
