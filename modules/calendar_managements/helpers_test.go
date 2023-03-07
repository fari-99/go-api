package calendar_managements

import (
	"testing"
	"time"
)

func TestGetBusinessDay(t *testing.T) {
	currentDate, _ := time.Parse("2006-01-02", "2023-01-09")
	countDay := 20
	holidayHelpers := NewHolidayHelpers()
	resultDate, totalBusinessDay := holidayHelpers.getBusinessDayFromCount(currentDate, countDay)

	timeCheck, _ := time.Parse("2006-01-02", "2023-02-06")
	if !resultDate.Equal(timeCheck) {
		t.Logf("result date := %s", resultDate.String())
		t.Logf("total lapsed day := %d", totalBusinessDay)
		t.Fail()
	}
}
