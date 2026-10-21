package common

// TODO consider similar pattern for observations.go for sync reasons
import (
	"strings"
	"sync"
)

type DNSTapirTag uint32

const (
	TagAnomalousQtypeDistribution DNSTapirTag = iota
	TagInternalTesting
	TagUndefined /* Must be last */
)

var lock sync.RWMutex
var tagToString = map[DNSTapirTag]string{
	TagAnomalousQtypeDistribution: "anomalous_qtype_distribution",
	TagInternalTesting:            "internal_testing",
	TagUndefined:                  "undefined",
}

func (t DNSTapirTag) String() string {
	var tagStr string

	lock.RLock()
	if t >= TagUndefined {
		tagStr = tagToString[TagUndefined]
	} else {
		tagStr = tagToString[t]
	}
	lock.RUnlock()

	return tagStr
}

func StringToTag(str string) DNSTapirTag {
	tag := TagUndefined

	lock.RLock()
	for t, s := range tagToString { // TODO introduce reverse map instead?
		if s == strings.ToLower(str) {
			tag = t
			break
		}
	}
	lock.RUnlock()

	return tag
}
