package strtime

import (
	"sort"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

// #include <stdlib.h>
// #include "bsdshim.h"
// extern int bsd_strptime(const char *s, const char *format, struct mytm *tm);
import "C"

// StringByLength 实现 sort.Interface 接口
type StringByLength []string

func (a StringByLength) Len() int {
	return len(a)
}

func (a StringByLength) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a StringByLength) Less(i, j int) bool {
	// 按照长度从大到小排序
	return len(a[i]) > len(a[j])
}

// 定义一些常用的日期时间格式
var defaultLayouts = []string{
	"%Y-%m-%d %H:%M:%S",
	"%Y-%m-%dT%H:%M:%S",
	"%Y-%m-%dT%H:%M:%SZ",
	"%Y-%m-%dT%H:%M:%S.%LZ",
	"%Y-%m-%dT%H:%M:%S.%L",
	"%Y-%m-%dT%H:%M:%S.%L%z",
	"%Y-%m-%d",
	"%Y/%m/%d %H:%M:%S",
	"%Y/%m/%dT%H:%M:%S",
	"%Y/%m/%d",
	"%Y%m%d %H:%M:%S",
	"%Y%m%dT%H:%M:%S",
	"%Y%m%d",
	"%a %d %b %Y %H:%M:%S",
	"%A %d %B %Y %H:%M:%S",
	"%d %b %Y %H:%M:%S",
	"%d %B %Y %H:%M:%S",
	"%d %b %Y %I:%M:%S %p",
	"%d %B %Y %I:%M:%S %p",
	"%d %b %Y %l:%M:%S %p",
	"%d %B %Y %l:%M:%S %p",
	"%d %b %Y %H:%M",
	"%d %B %Y %H:%M",
	"%d %b %Y %I:%M %p",
	"%d %B %Y %I:%M %p",
	"%d %b %Y %l:%M %p",
	"%d %B %Y %l:%M %p",
	"%d %b %Y %H:%M:%S %Z",
	"%d %B %Y %H:%M:%S %Z",
	"%d %b %Y %I:%M:%S %p %Z",
	"%d %B %Y %I:%M:%S %p %Z",
	"%d %b %Y %l:%M:%S %p %Z",
	"%d %B %Y %l:%M:%S %p %Z",
	"%d %b %Y %H:%M %Z",
	"%d %B %Y %H:%M %Z",
	"%d %b %Y %I:%M %p %Z",
	"%d %B %Y %I:%M %p %Z",
	"%d %b %Y %l:%M %p %Z",
	"%d %B %Y %l:%M %p %Z",
	"%Y-%m-%d %H:%M:%S %Z",
	"%Y-%m-%d %H:%M:%S %z",
	"%Y-%m-%d %I:%M:%S %p %Z",
	"%Y-%m-%d %I:%M:%S %p %z",
	"%Y-%m-%d %l:%M:%S %p %Z",
	"%Y-%m-%d %l:%M:%S %p %z",
	"%Y-%m-%d %H:%M %Z",
	"%Y-%m-%d %H:%M %z",
	"%Y-%m-%d %I:%M %p %Z",
	"%Y-%m-%d %I:%M %p %z",
	"%Y-%m-%d %l:%M %p %Z",
	"%Y-%m-%d %l:%M %p %z",
	"%Y-%m-%d %H:%M:%S %Z %z",
	"%Y-%m-%d %I:%M:%S %p %Z %z",
	"%Y-%m-%d %l:%M:%S %p %Z %z",
	"%Y-%m-%d %H:%M %Z %z",
	"%Y-%m-%d %I:%M %p %Z %z",
	"%Y-%m-%d %l:%M %p %Z %z",
	"%Y-%m-%d %H:%M:%S %z %Z",
	"%Y-%m-%d %I:%M:%S %p %z %Z",
	"%Y-%m-%d %l:%M:%S %p %z %Z",
	"%Y-%m-%d %H:%M %z %Z",
	"%Y-%m-%d %I:%M %p %z %Z",
	"%Y-%m-%d %l:%M %p %z %Z",
	"%Y-%m-%d %H:%M:%S %z %z",
	"%Y-%m-%d %I:%M:%S %p %z %z",
	"%Y-%m-%d %l:%M:%S %p %z %z",
	"%Y-%m-%d %H:%M %z %z",
	"%Y-%m-%d %I:%M %p %z %z",
	"%Y-%m-%d %l:%M %p %z %z",
	"%Y-%m-%d %H:%M:%S %Z %z %Z",
	"%Y-%m-%d %I:%M:%S %p %Z %z %Z",
	"%Y-%m-%d %l:%M:%S %p %Z %z %Z",
	"%Y-%m-%d %H:%M %Z %z %Z",
	"%Y-%m-%d %I:%M %p %Z %z %Z",
	"%Y-%m-%d %l:%M %p %Z %z %Z",
	"%Y-%m-%d %H:%M:%S %z %z %Z",
	"%Y-%m-%d %I:%M:%S %p %z %z %Z",
	"%Y-%m-%d %l:%M:%S %p %z %z %Z",
	"%Y-%m-%d %H:%M %z %z %Z",
	"%Y-%m-%d %I:%M %p %z %z %Z",
	"%Y-%m-%d %l:%M %p %z %z %Z",
	"%Y-%m-%d %H:%M:%S %Z %z %z",
	"%Y-%m-%d %I:%M:%S %p %Z %z %z",
	"%Y-%m-%d %l:%M:%S %p %Z %z %z",
	"%Y-%m-%d %H:%M %Z %z %z",
	"%Y-%m-%d %I:%M %p %Z %z %z",
	"%Y-%m-%d %l:%M %p %Z %z %z",
	"%Y-%m-%d %H:%M:%S %z %z %Z",
	"%Y-%m-%d %I:%M:%S %p %z %z %Z",
	"%Y-%m-%d %l:%M:%S %p %z %z %Z",
	"%Y-%m-%d %H:%M %z %z %Z",
	"%Y-%m-%d %I:%M %p %z %z %Z",
	"%Y-%m-%d %l:%M %p %z %z %Z",
	"%B %dth, %Y",             // July 4th, 2017
	"%B %dth, %Y %H:%M:%S %Z", // July 4th, 2017 12:39:30 BST
	"%B %dth, %Y %I%p",        // July 4th, 2017 11am
	"%B %dth, %Y %I%p",        // July 4th, 2017 12pm
	"%m/%d/%y",                // 7/4/17
	"%d-%m-%Y",                // 04-07-2017
	"%Y-%b-%d %I%p",           // 2017-Jul-04 noon
	"%Y-%m-%d %H:%M:%S %Z%z",  // 2017-07-04 12:48:07 GMT+0545
	"%Y-%m-%d %H:%M:%S %Z%z",  // 2017-07-04 12:48:07 GMT-0200
}

func Strptime(value string, layout string) (time.Time, error) {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	var cTime C.struct_mytm
	if layout == "" {
		sort.Sort(StringByLength(defaultLayouts))
		for _, l := range defaultLayouts {
			cl := C.CString(l)
			if r, err := C.bsd_strptime(cValue, cl, &cTime); r != 0 && err == nil {
				return time.Date(
					int(cTime.tm_year)+1900,
					time.Month(cTime.tm_mon+1),
					int(cTime.tm_mday),
					int(cTime.tm_hour),
					int(cTime.tm_min),
					int(cTime.tm_sec),
					int(cTime.tm_nsec),
					time.FixedZone(C.GoString(cTime.tm_zone), int(cTime.tm_gmtoff)),
				), nil
			}
			C.free(unsafe.Pointer(cl))
		}
		return time.Time{}, errors.New("no suitable format found")
	}

	cLayout := C.CString(layout)
	defer C.free(unsafe.Pointer(cLayout))

	if r, err := C.bsd_strptime(cValue, cLayout, &cTime); r == 0 || err != nil {
		if r == 0 {
			err = errors.New("")
		}
		return time.Time{}, errors.Wrapf(err, "could not parse %s as %s", value, layout)
	}
	return time.Date(
		int(cTime.tm_year)+1900,
		time.Month(cTime.tm_mon+1),
		int(cTime.tm_mday),
		int(cTime.tm_hour),
		int(cTime.tm_min),
		int(cTime.tm_sec),
		int(cTime.tm_nsec),
		time.FixedZone(C.GoString(cTime.tm_zone), int(cTime.tm_gmtoff)),
	), nil
}
