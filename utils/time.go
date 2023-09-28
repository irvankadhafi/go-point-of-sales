package utils

import (
	"github.com/goodsign/monday"
	"log"
	"time"
)

// FormatToWesternIndonesianTime format to western indonesian time
// expected format: 12 April 2020 14:30 WIB
func FormatToWesternIndonesianTime(layout string, t *time.Time) string {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatal(err)
	}
	return monday.Format(t.In(location), layout, monday.LocaleIdID)
}
