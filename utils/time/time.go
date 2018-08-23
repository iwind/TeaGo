package timeutil

import (
	"time"
	"bytes"
	"strconv"
	"fmt"
)

var weekShortDays = [...]string{
	"Sun",
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
}

// 文档 http://php.net/manual/en/function.date.php
// 目前没有支持的 S, L, o, y, B, g, h, v, e, I, O, P, c, r, U
func Format(t time.Time, format string) string {
	buffer := bytes.NewBuffer([]byte{})

	for _, runeItem := range format {
		char := string(runeItem)
		switch char {
		case "Y": //2016
			buffer.WriteString(strconv.Itoa(t.Year()))
		case "m": //01-12
			buffer.WriteString(fmt.Sprintf("%02d", int(t.Month())))
		case "n": //1-12
			buffer.WriteString(strconv.Itoa(int(t.Month())))
		case "d": // 01-31
			buffer.WriteString(fmt.Sprintf("%02d", t.Day()))
		case "z": // 0 -365
			buffer.WriteString(strconv.Itoa(t.YearDay() - 1))
		case "j": //1-31
			buffer.WriteString(strconv.Itoa(t.Day()))
		case "H": //00-23
			buffer.WriteString(fmt.Sprintf("%02d", t.Hour()))
		case "G": //0-23
			buffer.WriteString(strconv.Itoa(t.Hour()))
		case "i": //00-59
			buffer.WriteString(fmt.Sprintf("%02d", t.Minute()))
		case "s": //00-59
			buffer.WriteString(fmt.Sprintf("%02d", t.Second()))
		case "A": //AM or PM
			if t.Hour() < 12 {
				buffer.WriteString("AM")
			} else {
				buffer.WriteString("PM")
			}
		case "a": //am or pm
			if t.Hour() < 12 {
				buffer.WriteString("am")
			} else {
				buffer.WriteString("pm")
			}
		case "u": //654321
			buffer.WriteString(strconv.Itoa(t.Nanosecond() / 1000))
		case "w": //weekday, 0, 1, 2, ...
			buffer.WriteString(strconv.Itoa(int(t.Weekday())))
		case "W": //ISO-8601 week，一年中第N周
			_, week := t.ISOWeek()
			buffer.WriteString(strconv.Itoa(week))
		case "N": //1, 2, ...7
			weekday := t.Weekday()
			if weekday == 0 {
				buffer.WriteString("7")
			} else {
				buffer.WriteString(strconv.Itoa(int(weekday)))
			}
		case "D": //Mon ... Sun
			buffer.WriteString(weekShortDays[t.Weekday()])
		case "l": //Monday ... Sunday
			buffer.WriteString(t.Weekday().String())
		case "t": //一个月中的天数
			t2 := time.Date(t.Year(), t.Month(), 32, 0, 0, 0, 0, time.Local)
			daysInMonth := 32 - t2.Day()

			buffer.WriteString(strconv.Itoa(daysInMonth))
		case "F": //January
			buffer.WriteString(t.Month().String())
		case "M": //Jan
			buffer.WriteString(t.Format("Jan"))
		case "T": //时区名，比如CST
			zone, _ := t.Zone()
			buffer.WriteString(zone)
		case "Z": //时区offset，比如28800
			_, offset := t.Zone()
			buffer.WriteString(strconv.Itoa(offset))
		default:
			buffer.WriteRune(runeItem)
		}
	}

	return buffer.String()
}
