package common

// TODO consider similar pattern for observations.go for sync reasons
import (
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
	lock.RLock()
	tagStr := tagToString[t]

	if t >= TagUndefined {
		return tagToString[TagUndefined]
	}

	lock.RUnlock()
	return tagStr
}
