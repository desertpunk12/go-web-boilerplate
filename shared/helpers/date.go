package helpers

import (
	"fmt"
	"time"
)

func GetTodayDate() string {
	localLocationTime, _ := time.LoadLocation("Asia/Manila")
	today := time.Now().In(localLocationTime)
	return today.Format("2006-01-02")
}

func IsDateToday(date string) bool {
	loc, err := time.LoadLocation("Asia/Manila")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return false
	}
	todayRaw := time.Now().In(loc)
	today := todayRaw.Format("2006-01-02")
	return date == today
}
