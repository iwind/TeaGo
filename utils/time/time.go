package timeutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Format 文档 http://php.net/manual/en/function.date.php
// 目前没有支持的 S, L, o, B, v, e, I
func Format(format string, now ...time.Time) string {
	var t1 time.Time
	if len(now) > 0 {
		t1 = now[0]
	} else {
		t1 = time.Now()
	}

	buffer := strings.Builder{}

	for _, runeItem := range format {
		switch runeItem {
		case 'Y': // 年份，比如：2016
			buffer.WriteString(t1.Format("2006"))
		case 'y': // 年份后两位，比如：16
			buffer.WriteString(t1.Format("06"))
		case 'm': // 01-12
			buffer.WriteString(fmt.Sprintf("%02d", int(t1.Month())))
		case 'n': // 1-12
			buffer.WriteString(strconv.Itoa(int(t1.Month())))
		case 'd': // 01-31
			buffer.WriteString(fmt.Sprintf("%02d", t1.Day()))
		case 'z': // 0 -365
			buffer.WriteString(strconv.Itoa(t1.YearDay() - 1))
		case 'j': // 1-31
			buffer.WriteString(strconv.Itoa(t1.Day()))
		case 'H': // 00-23
			buffer.WriteString(fmt.Sprintf("%02d", t1.Hour()))
		case 'G': // 0-23
			buffer.WriteString(strconv.Itoa(t1.Hour()))
		case 'g': // 小时：1-12
			buffer.WriteString(t1.Format("3"))
		case 'h': // 小时：01-12
			buffer.WriteString(t1.Format("03"))
		case 'i': // 00-59
			buffer.WriteString(fmt.Sprintf("%02d", t1.Minute()))
		case 's': // 00-59
			buffer.WriteString(fmt.Sprintf("%02d", t1.Second()))
		case 'A': // AM or PM
			buffer.WriteString(t1.Format("PM"))
		case 'a': // am or pm
			buffer.WriteString(t1.Format("pm"))
		case 'u': // 微秒：654321
			buffer.WriteString(strconv.Itoa(t1.Nanosecond() / 1000))
		case 'v': // 毫秒：654
			buffer.WriteString(strconv.Itoa(t1.Nanosecond() / 1000000))
		case 'w': // weekday, 0, 1, 2, ...
			buffer.WriteString(strconv.Itoa(int(t1.Weekday())))
		case 'W': // ISO-8601 week，一年中第N周
			_, week := t1.ISOWeek()
			buffer.WriteString(strconv.Itoa(week))
		case 'N': // 1, 2, ...7
			weekday := t1.Weekday()
			if weekday == 0 {
				buffer.WriteString("7")
			} else {
				buffer.WriteString(strconv.Itoa(int(weekday)))
			}
		case 'D': // Mon ... Sun
			buffer.WriteString(t1.Format("Mon"))
		case 'l': // Monday ... Sunday
			buffer.WriteString(t1.Format("Monday"))
		case 't': // 一个月中的天数
			t2 := time.Date(t1.Year(), t1.Month(), 32, 0, 0, 0, 0, time.Local)
			daysInMonth := 32 - t2.Day()

			buffer.WriteString(strconv.Itoa(daysInMonth))
		case 'F': // January
			buffer.WriteString(t1.Format("January"))
		case 'M': // Jan
			buffer.WriteString(t1.Format("Jan"))
		case 'O': // 格林威治时间差（GMT），比如：+0800
			buffer.WriteString(t1.Format("-0700"))
		case 'P': // 格林威治时间差（GMT），比如：+08:00
			buffer.WriteString(t1.Format("-07:00"))
		case 'T': // 时区名，比如CST
			zone, _ := t1.Zone()
			buffer.WriteString(zone)
		case 'Z': // 时区offset，比如28800
			_, offset := t1.Zone()
			buffer.WriteString(strconv.Itoa(offset))
		case 'c': // ISO 8601，类似于：2004-02-12T15:19:21+00:00
			buffer.WriteString(t1.Format("2006-01-02T15:04:05Z07:00"))
		case 'r': // RFC 2822，类似于：Thu, 21 Dec 2000 16:01:07 +0200
			buffer.WriteString(t1.Format("Mon, 2 Jan 2006 15:04:05 -0700"))
		case 'U': // 时间戳
			buffer.WriteString(fmt.Sprintf("%d", t1.Unix()))
		default:
			buffer.WriteRune(runeItem)
		}
	}

	return buffer.String()
}

// FormatTime 格式化时间戳
func FormatTime(format string, timestamp int64) string {
	return Format(format, time.Unix(timestamp, 0))
}
