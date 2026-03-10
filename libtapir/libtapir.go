package libtapir

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/dnstapir/edm/pkg/protocols" // TODO replace with local implementation?

	"github.com/dnstapir/tapir-analyse-lib/common"
	"github.com/dnstapir/tapir-analyse-lib/logger"
)

const c_ID = "tal-libtapir"

type Conf struct {
	Log   common.Logger
	Debug bool `toml:"debug"`
}

type libtapir struct {
	id  string
	log common.Logger
}

func New(conf Conf) *libtapir {
	lt := new(libtapir)
	lt.id = c_ID
	if conf.Log == nil {
		log := logger.New(
			logger.Conf{
				Debug: conf.Debug,
			})
		lt.log = log
	} else {
		lt.log = conf.Log
	}
	defer lt.log.Debug("%s: debug logging enabled", lt.id)

	return lt
}

func (lt *libtapir) ExtractDomain(msgJson []byte) (string, error) {
	var newQnameEvent protocols.NewQnameJSON
	dec := json.NewDecoder(bytes.NewReader(msgJson))

	dec.DisallowUnknownFields()

	err := dec.Decode(&newQnameEvent)
	if err != nil {
		lt.log.Error("Error decoding qname from 'new qname event' msg")
		return "", err
	}

	return lt.NormalizeDomainName(newQnameEvent.Qname), nil
}

func (lt *libtapir) GenerateObservationMsg(domainStr string, flags uint32, ttl int) (string, error) {
	dom := domain{
		Name:         domainStr,
		TimeAdded:    time.Now(),
		TTL:          ttl,
		TagMask:      flags,
		ExtendedTags: []string{},
	}

	tapirMsg := tapirMsg{
		SrcName:   "dns-tapir",
		Creator:   "observation-encoder",
		MsgType:   "observation",
		ListType:  "doubtlist",
		Added:     []domain{dom},
		Removed:   []domain{},
		Msg:       "",
		TimeStamp: time.Now(),
		TimeStr:   "",
	}

	outMsg, err := json.Marshal(tapirMsg)
	if err != nil {
		lt.log.Error("Error serializing message, discarding...")
		return "", err
	}

	return string(outMsg), nil
}

func (lt *libtapir) ExtractObservations(data []byte) (map[string]uint32, error) {
	obs := make(map[string]uint32)
	var msg tapirMsg

	dec := json.NewDecoder(bytes.NewReader(data))
	if dec == nil {
		lt.log.Error("Problem creating decoder for json data")
		return nil, errors.New("bad json")
	}

	dec.DisallowUnknownFields()
	err := dec.Decode(&msg)
	if err != nil {
		lt.log.Error("Problem decoding JSON: %s", err)
		return nil, err
	}

	for _, d := range msg.Added {
		obs[lt.NormalizeDomainName(d.Name)] = d.TagMask
	}

	if len(obs) == 0 {
		lt.log.Error("Data contained no observations")
		return nil, errors.New("no observations found")
	}

	return obs, nil
}

func (lt *libtapir) NormalizeDomainName(name string) string {
	// make sure domain name ends with "." and is all lowercase
	nameLowered := strings.ToLower(name)
	nameTrimmed := strings.Trim(nameLowered, ".*")
	nameNormalized := nameTrimmed + "."

	return nameNormalized
}
