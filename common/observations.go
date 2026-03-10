package common

import (
	"time"
)

type Observation struct {
	Encoding uint32
	Bucket   string
	Ttl      time.Duration
}

const OBS_GLOBALLY_NEW string = "globally_new"
const OBS_GLOBALLY_NEW_ENC uint32 = 1
const OBS_GLOBALLY_NEW_TTL time.Duration = 7200 * time.Second

const OBS_LOOPTEST string = "looptest"
const OBS_LOOPTEST_ENC uint32 = 1024
const OBS_LOOPTEST_TTL time.Duration = 3600 * time.Second

const bucketSuffix string = "" // TODO use "_bucket"? Affects what we already have in NATS

var OBS_MAP = map[string]Observation{
	OBS_GLOBALLY_NEW: Observation{
		Encoding: OBS_GLOBALLY_NEW_ENC,
		Bucket:   OBS_GLOBALLY_NEW + bucketSuffix,
		Ttl:      OBS_GLOBALLY_NEW_TTL,
	},
	OBS_LOOPTEST: Observation{
		Encoding: OBS_LOOPTEST_ENC,
		Bucket:   OBS_LOOPTEST + bucketSuffix,
		Ttl:      OBS_LOOPTEST_TTL,
	},
}
