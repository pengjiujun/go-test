package util

import (
	"fmt"
	"time"
)

type LocalTime time.Time

func (t LocalTime) MarshalJSON() ([]byte, error) {
	// 格式化为：2006-01-02 15:04:05
	formatted := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(formatted), nil
}
