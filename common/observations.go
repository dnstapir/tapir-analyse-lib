package common

const OBS_GLOBALLY_NEW string = "globally_new"
const OBS_GLOBALLY_NEW_ENC uint32 = 1

const OBS_LOOPTEST string = "looptest"
const OBS_LOOPTEST_ENC uint32 = 1024

var OBS_MAP = map[string]uint32{
	OBS_GLOBALLY_NEW: OBS_GLOBALLY_NEW_ENC,
	OBS_LOOPTEST:     OBS_LOOPTEST_ENC,
}
