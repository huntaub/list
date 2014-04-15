package schedule

import (
	// "encoding/xml"
	"code.google.com/p/go.net/html"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func ParseList(ll io.Reader) (ClassList, error) {
	doc, err := html.Parse(ll)
	if err != nil {
		fmt.Println(err, "Couldn't parse Lou's List")
		return ClassList{}, err
	}

	foundTable := false
	var table *html.Node
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tbody" {
			foundTable = true
			table = n
		}
		if n.Data != "tbody" && !foundTable {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	f(doc)
	return parseTable(table), nil
}

func parseTable(table *html.Node) ClassList {
	// Loop Through Every Row of the Table
	output := ClassList{}
	var currentClass *Class = nil
	var currentSection *Section = &Section{}
	currentTimes, currentLocations, currentInstructors := []string{}, []string{}, []string{}
	for c := table.FirstChild; c != nil; c = c.NextSibling {
		for _, x := range c.Attr {
			if x.Key == "class" &&
				(strings.Contains(x.Val, "SectionOdd") || strings.Contains(x.Val, "SectionEven")) {
				// fmt.Print("\t\t")
				current := c.FirstChild
				for i := 0; i < 9; i++ {
					if i == 8 {
						// fmt.Print("\n")
						break
					}
					if current.FirstChild == nil {
						// fmt.Print("\t")
						if i == 2 {
							// fmt.Print("\t")
						}
					} else {
						if i == 0 && currentSection.Topic == "" && currentSection.Type != "" {
							currentSection.Meetings = ParseTime(currentTimes, currentLocations, currentInstructors)
							if currentSection.Number != 0 {
								currentClass.Sections = append(currentClass.Sections, currentSection)
							}
							currentSection = &Section{}
							currentTimes, currentLocations, currentInstructors = []string{}, []string{}, []string{}
						}
						if i == 0 {
							strSisNumber := strings.Trim(current.FirstChild.Data, " ")
							sisNumber, _ := strconv.Atoi(strSisNumber[0 : len(strSisNumber)-2]) // strange unicode characters at the end
							currentSection.SISNumber = sisNumber
							// fmt.Print(strings.Trim(current.FirstChild.Data, " "), "\t")
						} else if i == 1 {
							number, _ := strconv.Atoi(strings.Trim(current.FirstChild.Data, " "))
							currentSection.Number = number
							// fmt.Print(strings.Trim(current.FirstChild.Data, " "), "\t")
						} else if i == 2 {
							// Print Lecture Information
							if current.FirstChild.FirstChild == nil {
								// fmt.Print("\r")
								break
							}
							// Lecture...
							currentSection.Type = strings.Trim(current.FirstChild.FirstChild.Data, " ")
							// fmt.Print(strings.Trim(current.FirstChild.FirstChild.Data, " "), "\t")
							// Credits
							credits, _ := strconv.Atoi(strings.Trim(current.FirstChild.NextSibling.Data, " ()"))
							currentSection.Credits = credits
							// fmt.Print(strings.Trim(current.FirstChild.NextSibling.Data, " "), "\t")
						} else if i == 4 {
							// Print Enrollment Information
							enrollment := strings.Split(strings.Trim(current.FirstChild.FirstChild.Data, " "), "/")
							currentSection.Enrollment, _ = strconv.Atoi(strings.Trim(enrollment[0], " "))
							currentSection.Capacity, _ = strconv.Atoi(strings.Trim(enrollment[1], " "))
							// fmt.Print(strings.Trim(current.FirstChild.FirstChild.Data, " "), "\t")
						} else if i == 5 {
							// Print Instructor Information
							if current.FirstChild.FirstChild.Data != "span" {
								// Probably Staffff...
								currentInstructors = append(currentInstructors, strings.Trim(current.FirstChild.FirstChild.Data, " "))
								// fmt.Print(strings.Trim(current.FirstChild.FirstChild.Data, " "), "\t")
							} else {
								// A REAL INSTRUCTOR
								currentInstructors = append(currentInstructors, strings.Trim(current.FirstChild.FirstChild.FirstChild.Data, " "))
								// fmt.Print(strings.Trim(current.FirstChild.FirstChild.FirstChild.Data, " "), "\t")
							}
						} else if i == 6 {
							currentTimes = append(currentTimes, strings.Trim(current.FirstChild.Data, " "))
							// fmt.Print(strings.Trim(current.FirstChild.Data, " "), "\t")
						} else if i == 7 {
							// Location Information
							currentLocations = append(currentLocations, strings.Trim(current.FirstChild.Data, " "))
							// fmt.Print(strings.Trim(current.FirstChild.Data, " "), "\t")
						} else {
							// fmt.Print(strings.Trim(current.FirstChild.Data, " "), "\t")
						}
					}
					current = current.NextSibling
				}
				continue
			} else if x.Key == "class" && strings.Contains(x.Val, "SectionTopic") {
				currentSection.Meetings = ParseTime(currentTimes, currentLocations, currentInstructors)
				if currentSection.Number != 0 {
					currentClass.Sections = append(currentClass.Sections, currentSection)
				}
				currentSection = &Section{}
				currentTimes, currentLocations, currentInstructors = []string{}, []string{}, []string{}
				currentSection.Topic = strings.Trim(c.FirstChild.NextSibling.FirstChild.Data, " ")
				// fmt.Println("\t\t---", c.FirstChild.NextSibling.FirstChild.Data)
			}
		}

		// Get the TD Element
		theData := c.FirstChild
		if theData == nil {
			continue
		}
		for _, x := range theData.Attr {
			// This is the Beginning of the Department
			if x.Key == "class" && x.Val == "UnitName" {
				// fmt.Println(theData.FirstChild.Data)
			} else if x.Key == "class" && x.Val == "CourseName" {
				if currentClass != nil {
					currentSection.Meetings = ParseTime(currentTimes, currentLocations, currentInstructors)
					if currentSection.Number != 0 {
						currentClass.Sections = append(currentClass.Sections, currentSection)
					}
					currentSection = &Section{}
					currentTimes, currentLocations, currentInstructors = []string{}, []string{}, []string{}
					output.add(currentClass)
				}
				title := strings.Trim(theData.FirstChild.NextSibling.Data, " ")
				comp := strings.Split(title, " ")
				dept := strings.Split(comp[0], " ")[0]
				number, _ := strconv.Atoi(strings.Split(comp[0], " ")[1])
				name := ""
				if len(comp) == 2 {
					name = comp[1]
				}
				currentClass = &Class{}
				currentClass.Name = name
				currentClass.Number = number
				currentClass.Department = dept
				// fmt.Println("\t", dept, number, name)
			}
		}
	}
	// fmt.Println(output)
	return output
}

func ParseTime(timeString []string, locationString []string, instructorString []string) []SectionTime {
	output := []SectionTime{}
	for i, meet_time := range timeString {
		newSection := SectionTime{}
		if meet_time == "TBA" {
			newSection.TBA = true
			output = append(output, newSection)
			continue
		}

		var err error

		tempC := strings.Split(meet_time, " ")
		days := tempC[0]
		times := []string{tempC[1], tempC[3]}
		newSection.StartTime, err = time.Parse("3:04PM", times[0])
		if err != nil {
			fmt.Println("Couldn't parse start time. Try again.")
			continue
		}
		newSection.EndTime, err = time.Parse("3:04PM", times[1])
		if err != nil {
			fmt.Println("Couldn't parse end time. Try again.")
			continue
		}

		daysOfWeek := []time.Weekday{}
		for i, len := 0, len(days); i < len; i += 2 {
			dayOfWeek := string(days[i : i+2])

			switch dayOfWeek {
			case "Mo":
				daysOfWeek = append(daysOfWeek, time.Monday)
			case "Tu":
				daysOfWeek = append(daysOfWeek, time.Tuesday)
			case "We":
				daysOfWeek = append(daysOfWeek, time.Wednesday)
			case "Th":
				daysOfWeek = append(daysOfWeek, time.Thursday)
			case "Fr":
				daysOfWeek = append(daysOfWeek, time.Friday)
			case "Sa":
				daysOfWeek = append(daysOfWeek, time.Saturday)
			}
		}
		newSection.Days = daysOfWeek

		if len(locationString) != 0 {
			newSection.Location = locationString[i]
		}
		if len(instructorString) != 0 {
			newSection.Instructor = instructorString[i]
		}

		output = append(output, newSection)
	}
	return output
}
