package schedule

import (
	// "fmt"
	"strconv"
	"time"
)

type ClassList map[string]*Class

func (v ClassList) add(c *Class) {
	v[c.Department+" "+strconv.Itoa(c.Number)] = c
}

func (v ClassList) String() string {
	output := ""
	for _, x := range v {
		output += x.String()
	}
	return output
}

type Class struct {
	Name       string
	Department string
	Number     int
	Sections   SectionList
	SectionMap map[int]*Section
}

type Section struct {
	// Basic Information
	Number    int
	Type      string
	SISNumber int

	// Date Information
	Meetings Meetings

	// Enrollment Information
	Capacity   int
	Enrollment int

	// Extra Information
	Topic   string
	Credits int
}

func (s *Section) Open() bool {
	return (s.Enrollment < s.Capacity)
}

type SectionTime struct {
	Location   string
	Instructor string
	StartTime  time.Time
	EndTime    time.Time
	Days       MeetDays
	tba        bool
}

type Meetings []SectionTime

type MeetDays []time.Weekday

type SectionList []*Section

func (c *Class) CreateSectionMap() {
	if c.SectionMap == nil {
		c.SectionMap = make(map[int]*Section)
		for _, s := range c.Sections {
			c.SectionMap[s.Number] = s
		}
	}
}

func (c *Class) ValidClassTimes() []ClassTime {
	sMap := make(map[int][]*Section)
	hasLab, hasDiscussion := false, false
	for _, x := range c.Sections {
		switch x.Type {
		case "Discussion":
			hasDiscussion = true
		case "Laboratory":
			hasLab = true
		}
		current, ok := sMap[(x.Number / 100)]
		if !ok {
			current = []*Section{}
		}
		sMap[(x.Number / 100)] = append(current, x)
	}
	output := []ClassTime{}
	if hasLab {
		// Loop over each group of classes
		// With Lab classes, the first set of Sections are Lectures, and the Other Set are Labs
		lect := -1
		lab := -1
		for k, _ := range sMap {
			if lect == -1 {
				lect = k
			} else {
				lab = k
			}
		}
		for _, lectSect := range sMap[lect] {
			for _, labSect := range sMap[lab] {
				newCT := ClassTime{
					Class:         c,
					SectionNumber: []int{lectSect.Number, labSect.Number},
				}
				output = append(output, newCT)
			}
		}
	} else if hasDiscussion {
		// Loop over each group of classes
		lectures := []*Section{}
		capturedLectures := false
		// discussions := []*Section{}
		for _, v := range sMap {
			currentLectures := []*Section{}
			currentDiscussions := []*Section{}
			for _, x := range v {
				if x.Type == "Lecture" {
					currentLectures = append(currentLectures, x)
				} else if x.Type == "Discussion" {
					currentDiscussions = append(currentDiscussions, x)
				}
			}
			if len(currentLectures) > 0 && len(currentDiscussions) > 0 {
				for _, lectSect := range currentLectures {
					for _, discSect := range currentDiscussions {
						newCT := ClassTime{
							Class:         c,
							SectionNumber: []int{lectSect.Number, discSect.Number},
						}
						output = append(output, newCT)
					}
				}
			} else if len(currentDiscussions) == 0 {
				lectures = currentLectures
			} else if len(currentLectures) == 0 {
				capturedLectures = true
				for _, lectSect := range lectures {
					for _, discSect := range currentDiscussions {
						newCT := ClassTime{
							Class:         c,
							SectionNumber: []int{lectSect.Number, discSect.Number},
						}
						output = append(output, newCT)
					}
				}
			}
		}
		if !capturedLectures && len(lectures) > 0 {
			for _, lectSect := range lectures {
				newCT := ClassTime{
					Class:         c,
					SectionNumber: []int{lectSect.Number},
				}
				output = append(output, newCT)
			}
		}
	} else {
		for _, v := range sMap {
			for _, x := range v {
				newCT := ClassTime{
					Class:         c,
					SectionNumber: []int{x.Number},
				}
				output = append(output, newCT)
			}
		}
	}
	// fmt.Println(hasLecture, hasIndStudy, hasLab, hasDiscussion)
	return output
	// fmt.Println(sMap)
}

func (c *Class) Equals(b *Class) bool {
	return (c.Department == b.Department) && (c.Number == b.Number)
}

func (c *Class) String() string {
	output := "Class "
	if c.Number >= 6000 {
		output += "(G) "
	}
	return output + c.Department + " " + strconv.Itoa(c.Number) + ": " + c.Name + "\n" + c.Sections.String()
}

func (c *Class) PrintForCalendar(line int, sectionIndex int, meetingIndex int) string {
	output := c.Department + " " + strconv.Itoa(c.Number)
	// theSection := c.Sections[sectionIndex]
	// sectionLines := theSection.Meetings[meetingIndex].LinesToPrint()

	// if line == 1 && theSection.Location != "" {
	// output = "SEC " + strconv.Itoa(theSection.Number) + " " + theSection.Location
	// } else if line != 0 && line < sectionLines {
	i := len(output)
	output = ""
	for j := 0; j < i; j += 1 {
		output += "."
	}
	// }
	return output
}

func (c *Class) Descriptor() string {
	return c.Department + " " + strconv.Itoa(c.Number)
}

func (s Meetings) String() string {
	output := ""
	for _, x := range s {
		output += "(Meets " + x.Days.String() + " from " + x.StartTime.Format("3:04 PM") + " to " + x.EndTime.Format("3:04 PM") + " in " + x.Location + ") "
	}
	return output
}

func (s *SectionTime) LinesToPrint() int {
	return int(s.EndTime.Sub(s.StartTime) / time.Minute / 10)
}

func (s *Section) String() string {
	if s.Topic != "" {
		return s.Type + " " + strconv.Itoa(s.Number) + " (" + s.Topic + ") Meets " + s.Meetings.String()
	} else {
		return s.Type + " " + strconv.Itoa(s.Number) + " Meets " + s.Meetings.String()
	}
}

func (s *Section) Overlaps(b *Section) bool {
	if b == nil || s == nil {
		return false
	}
	for _, x := range s.Meetings {
		for _, y := range b.Meetings {
			if x.Overlaps(y) {
				return true
			}
		}
	}
	return false
}

func (s SectionTime) Overlaps(b SectionTime) bool {
	if !s.Days.Overlaps(b.Days) {
		return false
	}
	return (TimeBeforeOrEqualTo(s.StartTime, b.EndTime) && TimeAfterOrEqualTo(s.StartTime, b.StartTime)) ||
		(TimeBeforeOrEqualTo(b.StartTime, s.EndTime) && TimeAfterOrEqualTo(b.StartTime, s.StartTime)) ||
		(TimeBeforeOrEqualTo(b.EndTime, s.EndTime) && TimeAfterOrEqualTo(b.EndTime, s.StartTime)) ||
		(TimeBeforeOrEqualTo(s.EndTime, b.EndTime) && TimeAfterOrEqualTo(s.EndTime, b.StartTime))
}

func (s SectionList) String() string {
	output := ""
	for _, v := range s {
		output += "\t" + v.String() + "\n"
	}
	return output
}

func (m MeetDays) Overlaps(b MeetDays) bool {
	for _, v := range b {
		if m.Contains(v) {
			return true
		}
	}
	return false
}

func (m MeetDays) String() string {
	output := ""
	for i, v := range m {
		if i > 0 {
			output += ", "
		}
		output += v.String()
	}
	return output
}

func (m MeetDays) ContainsString(a string) bool {
	var dayOfWeek time.Weekday
	switch a {
	case "M":
		dayOfWeek = time.Monday
	case "T":
		dayOfWeek = time.Tuesday
	case "W":
		dayOfWeek = time.Wednesday
	case "R":
		dayOfWeek = time.Thursday
	case "F":
		dayOfWeek = time.Friday
	case "S":
		dayOfWeek = time.Saturday
	}
	return m.Contains(dayOfWeek)
}

func (m MeetDays) Contains(a time.Weekday) bool {
	for _, v := range m {
		if v == a {
			return true
		}
	}
	return false
}
