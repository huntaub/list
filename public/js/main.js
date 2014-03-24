function createEvent(title, calStart, classStart, classEnd, classDays) {
	output = new Array();
	console.log(classStart)
	classStart = moment(classStart, "---, DD MMM YYYY HH:mm:ss EST");
	classEnd = moment(classEnd, "---, DD MMM YYYY HH:mm:ss EST");
	for(i = 0; i < 5; i++) {
		startTime = new Date(calStart);
		startTime.setHours(classStart.hour());
		startTime.setMinutes(classStart.minute());
		endTime = new Date(calStart);
		endTime.setHours(classEnd.hour());
		endTime.setMinutes(classEnd.minute());
		if(i == 0 && classDays.indexOf("Monday") != -1) {
			output.push({
				title: title,
				start: startTime,
				end: endTime,
				allDay: false,
			});
		} else if (i == 1 && classDays.indexOf("Tuesday") != -1) {
			output.push({
				title: title,
				start: startTime,
				end: endTime,
				allDay: false,
			});
		} else if (i == 2 && classDays.indexOf("Wednesday") != -1) {
			output.push({
				title: title,
				start: startTime,
				end: endTime,
				allDay: false,
			});
		} else if (i == 3 && classDays.indexOf("Thursday") != -1) {
			output.push({
				title: title,
				start: startTime,
				end: endTime,
				allDay: false,
			});
		} else if (i == 4 && classDays.indexOf("Friday") != -1) {
			output.push({
				title: title,
				start: startTime,
				end: endTime,
				allDay: false,
			});
		}
		calStart = new Date (calStart.getTime() + 24 * 3600 * 1000)
		// calStart.setDay(calStart.getDay() + 1);
	}
	return output
}