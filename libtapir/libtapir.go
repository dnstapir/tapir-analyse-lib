package libtapir

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/dnstapir/edm/pkg/protocols" // TODO replace with local implementation?

	"github.com/dnstapir/tapir-analyse-lib/common"
)

func ExtractDomain(msgJson []byte) (string, error) {
	var newQnameEvent protocols.NewQnameJSON

	dec := json.NewDecoder(bytes.NewReader(msgJson))
	dec.DisallowUnknownFields()

	err := dec.Decode(&newQnameEvent)
	if err != nil {
		return "", err
	}

	_, err = dec.Token()
	if err != io.EOF {
		return "", common.ErrBadJSON
	}

	return NormalizeDomainName(newQnameEvent.Qname), nil
}

func GenerateObservationMsg(domainStr string, flags uint32, ttl int) (string, error) {
	if ttl <= 0 {
		return "", common.ErrBadParam
	}
	dom := common.FlaggedDomain{
		Name:         domainStr,
		TimeAdded:    time.Now(),
		TTL:          ttl,
		TagMask:      flags,
		ExtendedTags: []string{},
	}

	tapirMsg := common.TapirObs{
		SrcName:   "dns-tapir",
		Creator:   "observation-encoder",
		MsgType:   "observation",
		ListType:  "doubtlist",
		Added:     []common.FlaggedDomain{dom},
		Removed:   []common.FlaggedDomain{},
		Msg:       "",
		TimeStamp: time.Now(),
		TimeStr:   "",
	}

	outMsg, err := json.Marshal(tapirMsg)
	if err != nil {
		return "", err
	}

	return string(outMsg), nil
}

func ExtractObservations(data []byte) (map[string]uint32, error) {
	obs := make(map[string]uint32)
	var msg common.TapirObs

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	err := dec.Decode(&msg)
	if err != nil {
		return nil, err
	}

	_, err = dec.Token()
	if err != io.EOF {
		return nil, common.ErrBadJSON
	}

	for _, d := range msg.Added {
		name := NormalizeDomainName(d.Name)
		obs[name] |= d.TagMask
	}

	if len(obs) == 0 {
		return nil, errors.New("no observations found")
	}

	return obs, nil
}

func NormalizeDomainName(name string) string {
	/* make sure domain name ends with "." and is all lowercase */
	nameLowered := strings.ToLower(name)
	nameTrimmed := strings.Trim(nameLowered, ".*")
	nameNormalized := nameTrimmed + "."

	return nameNormalized
}

func NormalizeDomainNameSuffix(suffix string) string {
	/* make sure domain name starts and ends with "." and is all lowercase */
	normalized := NormalizeDomainName(suffix)
	if normalized == "." {
		return "."
	}

	return "." + normalized
}

func FlipDomainName(domain string) string {
	norm := NormalizeDomainName(domain)
	labels := strings.Split(norm, ".")
	nonEmpty := make([]string, 0, len(labels))
	for _, l := range labels {
		if l != "" {
			nonEmpty = append(nonEmpty, l)
		}
	}
	if len(nonEmpty) == 0 {
		return "."
	}
	slices.Reverse(nonEmpty)
	return strings.Join(nonEmpty, ".")
}
