// pkg/utils/timeutil/timeutil.go 
package timeutil

import "time"

// CalculateEDD calculates an estimated delivery date by adding a number of business days.
// It skips weekends (Saturday and Sunday).
func CalculateEDD(startTime time.Time, businessDays int) time.Time {
	edd := startTime
	for d := 0; d < businessDays; {
		edd = edd.AddDate(0, 0, 1) // Add one day
		weekday := edd.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			d++ // Only increment the counter if it's a weekday
		}
	}
	return edd
}