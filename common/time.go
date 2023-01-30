package common

import (
	"database/sql/driver"
	"errors"
	"time"
)

type Time time.Time

const (
	timeFormat = "2006-01-02 15:04:05"
)

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	*t = Time(now)
	return
}

func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(timeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, timeFormat)
	b = append(b, '"')
	return b, nil
}

func (t Time) String() string {
	return time.Time(t).Format(timeFormat)
}

func (t Time) Value() (driver.Value, error) {
	// MyTime 转换成 time.Time 类型
	tTime := time.Time(t)
	return tTime.Format(timeFormat), nil
}

func (t *Time) Scan(v interface{}) error {
	switch vt := v.(type) {
	case string:
		// 字符串转成 time.Time 类型
		tTime, _ := time.Parse(timeFormat, vt)
		*t = Time(tTime)
	case time.Time:
	case Time:
	default:
		return errors.New("类型处理错误")
	}
	return nil
}
