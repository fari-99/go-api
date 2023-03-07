package calendar_managements

import (
	"time"

	"go-api/constant"
	"go-api/modules/models"
)

type HolidayHelpers struct {
	holidays   map[string]int8
	countDays  int64
	targetDays time.Time
}

func NewHolidayHelpers() *HolidayHelpers {
	helper := HolidayHelpers{
		holidays: make(map[string]int8), // set empty map holiday
	}

	return &helper
}

func (h *HolidayHelpers) SetHolidays(holidays []models.CalendarManagements) *HolidayHelpers {
	var holidayMap map[string]int8
	for _, holiday := range holidays {
		holidayDateString := holiday.Dates.String()
		if _, ok := holidayMap[holidayDateString]; !ok {
			holidayMap[holidayDateString] = holiday.Status
		}
	}

	return h
}

func (h *HolidayHelpers) getBusinessDayFromCount(currentDate time.Time, n int) (resultDate *time.Time, totalBusinessDay int) {
	saturday := time.Saturday
	sunday := time.Sunday
	holidays := h.holidays

	totalHoliday := 0
	currentDateTmp := currentDate
	for i := 0; i < n; i++ {
		checkDate := currentDateTmp.AddDate(0, 0, 1) // adding 1 day at a time
		currentDateTmp = checkDate                   // update currentDateTmp to checked date

		checkDateDay := checkDate.Weekday()
		if checkDateDay == saturday || checkDateDay == sunday {
			totalHoliday += 1
			continue
		}

		if status, exists := holidays[checkDate.String()]; exists && status == constant.StatusActive {
			totalHoliday += 1
			continue
		}
	}

	if totalHoliday > 0 {
		newResult, newTotalBusinessDay := h.getBusinessDayFromCount(currentDateTmp, totalHoliday)
		return newResult, n + newTotalBusinessDay
	} else {
		return &currentDateTmp, n + totalHoliday
	}
}
