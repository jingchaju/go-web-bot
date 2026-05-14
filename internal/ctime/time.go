package ctime

import "time"

var ChinaLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

func Now() time.Time    { return time.Now().In(ChinaLocation) }
func Timestamp() int64  { return Now().Unix() }
func FormatNow() string { return Now().Format("2006-01-02 15:04:05") }
