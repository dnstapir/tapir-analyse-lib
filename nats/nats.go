package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/dnstapir/tapir-analyse-lib/common"
	"github.com/dnstapir/tapir-analyse-lib/libtapir"
	"github.com/dnstapir/tapir-analyse-lib/logger"
)

type Conf struct {
	Debug                    bool                `toml:"debug"`
	Url                      string              `toml:"url"`
	EventSubject             string              `toml:"event_subject"`
	ObservationSubjectPrefix string              `toml:"observation_subject_prefix"`
	ObservationBuckets       []ObservationBucket `toml:"observation_buckets"`
	SeenDomainsBucket        Bucket              `toml:"seen_domains_bucket"`
	SeenDomainsSubjectPrefix string              `toml:"seen_domains_subject_prefix"`
	SubjectSouthbound        string              `toml:"subject_southbound"`
	PrivateSubjectPrefix     string              `toml:"private_subject_prefix"`
	PrivateBucket            Bucket              `toml:"private_bucket"`
	AnalystID                string
	Log                      common.Logger
}

type ObservationBucket struct {
	Bucket
	Observation string `toml:"observation"`
}

type Bucket struct {
	Name   string `toml:"name"`
	Create bool   `toml:"create"`
	Ttl    int    `toml:"ttl"`
}

type natsClient struct {
	analystID                string
	log                      common.Logger
	url                      string
	conn                     *nats.Conn
	js                       jetstream.JetStream
	eventSubject             string
	okvMap                   obsKvMap
	observationSubjectPrefix string
	kvSeenDomains            jetstream.KeyValue
	seenDomainsSubjectPrefix string
	subjectSouthbound        string
	kvPrivate                jetstream.KeyValue // TODO currently unused
	privateSubjectPrefix     string
	watchersWG               sync.WaitGroup
}

type obsKvMap struct {
	sync.RWMutex // Needed? We only ever write during startup
	m            map[string]jetstream.KeyValue
}

func Create(conf Conf) (*natsClient, error) {
	nc := new(natsClient)

	if conf.Log == nil {
		log := logger.New(
			logger.Conf{
				Debug: conf.Debug,
			})
		nc.log = log
	} else {
		nc.log = conf.Log
	}
	nc.log.Debug("NATS debug logging enabled")

	nc.analystID = conf.AnalystID
	if nc.analystID == "" {
		nc.log.Info("No analyst identifier  configured. Functionality will be limited.")
	}

	if conf.Url == "" {
		nc.log.Error("Bad URL when creating NATS client")
		return nil, common.ErrBadParam
	}
	nc.url = conf.Url

	conn, err := nats.Connect(nc.url)
	if err != nil {
		nc.log.Error("Could not connect to NATS: %s", err)
		return nil, err
	}
	nc.conn = conn

	defer func() {
		if err != nil && nc.conn != nil {
			nc.conn.Close()
		}
	}()

	js, err := jetstream.New(nc.conn)
	if err != nil {
		nc.log.Error("Could not get a jetstream handle: %s", err)
		return nil, err
	}
	nc.js = js

	if conf.EventSubject == "" {
		nc.log.Info("Configuration for event subscription subject missing, will not subscribe")
	}
	nc.eventSubject = common.NormalizeNatsSubject(conf.EventSubject)

	if conf.PrivateSubjectPrefix == "" || conf.PrivateBucket.Name == "" {
		nc.log.Info("Configuration for private bucket missing, will not create")
	} else {
		nc.privateSubjectPrefix = common.NormalizeNatsSubjectPrefix(conf.PrivateSubjectPrefix)

		kvPriv, err := nc.setupBucket(conf.PrivateBucket)
		if err != nil {
			nc.log.Error("Could not initialize private bucket in NATS")
			return nil, err
		}
		nc.kvPrivate = kvPriv
	}

	if conf.SeenDomainsSubjectPrefix == "" || conf.SeenDomainsBucket.Name == "" {
		nc.log.Info("Configuration for seen domains bucket missing, will not create")
	} else {
		nc.seenDomainsSubjectPrefix = common.NormalizeNatsSubjectPrefix(conf.SeenDomainsSubjectPrefix)
		kvSeenDom, err := nc.setupBucket(conf.SeenDomainsBucket)
		if err != nil {
			nc.log.Error("Could not initialize seen domains bucket in NATS")
			return nil, err
		}
		nc.kvSeenDomains = kvSeenDom
	}

	nc.subjectSouthbound = strings.Trim(conf.SubjectSouthbound, common.NATS_DELIM)
	if nc.subjectSouthbound == "" {
		nc.log.Info("No southbound subject configured. Functionality will be limited.")
	}

	nc.okvMap.Lock()
	nc.okvMap.m = make(map[string]jetstream.KeyValue)
	nc.okvMap.Unlock()

	if conf.ObservationSubjectPrefix == "" {
		nc.log.Error("Bad observation subject prefix when creating NATS client")
		return nil, common.ErrBadParam
	}
	nc.observationSubjectPrefix = common.NormalizeNatsSubjectPrefix(conf.ObservationSubjectPrefix)

	err = nc.initObservationBuckets(conf.ObservationBuckets)
	if err != nil {
		nc.log.Error("Could not initialize observation buckets in NATS")
		return nil, err
	}

	return nc, nil
}

func (nc *natsClient) ActivateSubscription(ctx context.Context) (<-chan common.NatsMsg, error) {
	if nc.eventSubject == "" {
		nc.log.Error("No event subject configured, cannot activate subscription")
		return nil, common.ErrNotCompleted
	}

	rawChan := make(chan *nats.Msg, 100) // TODO adjustable buffer?
	sub, err := nc.conn.ChanSubscribe(nc.eventSubject, rawChan)
	if err != nil {
		nc.log.Error("Couldn't subscribe to raw nats channel: '%s'", err)
		return nil, err
	}

	outCh := make(chan common.NatsMsg, 100) // TODO adjustable buffer?
	go func() {
		defer close(outCh)
		defer func() { _ = sub.Unsubscribe() }()
		nc.log.Info("Starting NATS listener loop")
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-rawChan:
				if !ok {
					nc.log.Warning("Incoming NATS channel closed")
					return
				}
				nc.log.Debug("Incoming NATS message on '%s'!", msg.Subject)
				natsMsg := common.NatsMsg{
					Headers: make(map[string]string),
					Data:    msg.Data,
					Subject: msg.Subject,
				}
				for h, v := range msg.Header {
					if slices.Contains(common.NATSHEADERS_DNSTAPIR_ALL, h) {
						natsMsg.Headers[h] = v[0] // TODO use entire slice?
					}
				}
				select {
				case outCh <- natsMsg:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	nc.log.Info("Subscribed to '%s'", nc.eventSubject)

	return outCh, nil
}

func (nc *natsClient) Shutdown() error {
	nc.watchersWG.Wait()
	if nc.conn != nil {
		nc.conn.Close()
	}
	return nil
}

func (nc *natsClient) SetObservation(ctx context.Context, domain, obs string) error {
	if nc.analystID == "" {
		nc.log.Error("Attempted to set observation without having an analyst ID configured")
		return common.ErrNotCompleted
	}

	nc.okvMap.RLock()
	kv, ok := nc.okvMap.m[obs]
	nc.okvMap.RUnlock()
	if !ok {
		nc.log.Error("No bucket configured for observation '%s'", obs)
		return common.ErrBadParam
	}

	flipped := libtapir.FlipDomainName(domain)
	subject := strings.Join(
		[]string{
			nc.observationSubjectPrefix,
			obs,
			flipped,
		},
		common.NATS_DELIM)

	_, err := kv.Create(ctx, subject, []byte(nc.analystID))
	if errors.Is(err, jetstream.ErrKeyExists) {
		_, err := kv.Put(ctx, subject, []byte(nc.analystID))
		if err != nil {
			nc.log.Error("Couldn't set key '%s': '%s'", subject, err)
			return err
		}
	} else if err == nil {
		nc.log.Debug("Observation '%s' set for domain '%s'", obs, domain)
	} else {
		nc.log.Error("Couldn't create key '%s': '%s'", subject, err)
		return err
	}

	return nil
}

func (nc *natsClient) AddDomain(ctx context.Context, domain string, reporter string) (bool, error) {
	if nc.kvSeenDomains == nil {
		nc.log.Error("No NATS KV store configured for keeping track of seen domains")
		return false, common.ErrNotCompleted
	}

	subject := _getSubjectFromFqdn(domain, nc.seenDomainsSubjectPrefix, "")

	entry, err := nc.kvSeenDomains.Get(ctx, subject)
	if err != nil && !errors.Is(err, jetstream.ErrKeyNotFound) {
		nc.log.Error("Error accessing storage: %s, subject: %s", err, subject)
		return false, err
	}

	var updatedData []byte
	found := false
	timestamp := time.Now().Unix()
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		nc.log.Debug("Previously unseen domain '%s' by '%s'", domain, reporter)
		firstReport := make(map[string]int64)
		firstReport[reporter] = timestamp
		updatedData, err = json.Marshal(firstReport)
		if err != nil {
			nc.log.Error("Couldn't serialize first report: %s", err)
			return found, err
		}
		_, err = nc.kvSeenDomains.Put(ctx, subject, updatedData)
		if err != nil {
			nc.log.Warning("Couldn't update report history: %s", err)
		}
	} else if err == nil {
		found = true
		data := entry.Value()
		var reportHistory map[string]int64
		err = json.Unmarshal(data, &reportHistory)
		if err != nil {
			nc.log.Error("Couldn't read report history: %s", err)
			return found, err
		}

		_, ok := reportHistory[reporter]
		if !ok {
			reportHistory[reporter] = timestamp
		} else {
			nc.log.Debug("%s has already reported %s as seen", reporter, domain)
		}
		updatedData, err = json.Marshal(reportHistory)
		if err != nil {
			nc.log.Error("Couldn't serialize updated report: %s", err)
			return found, err
		}
		_, err := nc.kvSeenDomains.Update(ctx, subject, updatedData, entry.Revision())
		if err != nil {
			nc.log.Error("Couldn't update report history: %s", err)
			return found, err
		}
	} else {
		return found, err
	}

	return found, nil
}

func (nc *natsClient) WatchObservations(ctx context.Context) (<-chan common.NatsMsg, error) {
	aggrKeyChan := make(chan jetstream.KeyValueEntry, 128) // TODO adjustable buffer size?
	outCh := make(chan common.NatsMsg, 32)                 // TODO adjustable buffer size?
	atLeastOneBucket := false

	nc.okvMap.RLock()
	okvm := nc.okvMap.m
	nc.okvMap.RUnlock()
	for _, kv := range okvm {
		subjectParts := []string{nc.observationSubjectPrefix, common.NATS_GLOB}
		subject := strings.Join(subjectParts, common.NATS_DELIM)
		w, err := kv.Watch(ctx, subject, jetstream.UpdatesOnly())
		if err != nil {
			nc.log.Error("Couldn't watch bucket %s: '%s'. Skipping...", kv.Bucket(), err)
			continue
		} else {
			atLeastOneBucket = true
		}

		nc.watchersWG.Go(func() {
			for k := range w.Updates() {
				select {
				case aggrKeyChan <- k:
				case <-ctx.Done():
					return
				}
			}
			nc.log.Info("Watcher for bucket %s stopping", kv.Bucket())
		})

		nc.log.Info("Watching bucket '%s'", kv.Bucket())
	}

	nc.watchersWG.Go(func() {
		defer close(outCh)
		defer nc.log.Info("Leaving aggregate NATS listener loop")
		nc.log.Info("Starting aggregate NATS listener loop")
		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-aggrKeyChan:
				if !ok {
					nc.log.Warning("Aggregate NATS channel closed unexpectedly")
					return
				}
				if val == nil {
					continue
				}
				nc.log.Debug("Incoming NATS KV update on '%s'!", val.Key())
				natsMsg := common.NatsMsg{
					Headers: nil,
					Data:    val.Value(),
					Subject: val.Key(),
				}
				select {
				case outCh <- natsMsg:
				case <-ctx.Done():
					return
				}
			}
		}
	})

	if !atLeastOneBucket {
		return nil, errors.New("failed to watch any buckets")
	}

	return outCh, nil
}

func (nc *natsClient) GetObservations(ctx context.Context, domain string) (uint32, int, error) {
	var obsVec uint32
	ttl := math.MaxInt64

	nc.okvMap.RLock()
	defer nc.okvMap.RUnlock()

	nc.okvMap.RLock()
	okvm := nc.okvMap.m
	nc.okvMap.RUnlock()
	for obsStr, kv := range okvm {
		subject := nc.genObservationSubject(domain, obsStr)

		obsUint, ok := common.OBS_MAP[obsStr]
		if !ok {
			nc.log.Error("Unknown observation %s, will not look up in NATS", obsStr)
			continue
		}

		status, err := kv.Status(ctx)
		if err != nil {
			nc.log.Error("Couldn't get bucket status for %s: %s", kv.Bucket(), err)
			continue
		}
		ttlBucket := status.TTL()

		val, err := kv.Get(ctx, subject)
		if err == jetstream.ErrKeyNotFound {
			nc.log.Debug("No %s observations for %s", obsStr, domain)
			continue
		} else if err != nil {
			nc.log.Error("Couldn't get observation value for key %s: %s", subject, err)
			continue
		}

		keyAge := time.Now().Sub(val.Created())
		keyTtl := int(math.Round((ttlBucket - keyAge).Seconds()))
		if keyTtl <= 0 {
			nc.log.Warning("Bucket %s seems to contain outdated key %s", kv.Bucket(), subject)
			continue
		}

		obsVec |= obsUint
		if keyTtl < ttl {
			ttl = keyTtl
		}
	}

	/* If not updated, set to zero because something probably went bad and results are unreliable */
	if ttl == math.MaxInt64 {
		ttl = 0
		obsVec = 0
	}

	return obsVec, ttl, nil
}

func (nc *natsClient) SendSouthboundObservation(msg string) error {
	if nc.subjectSouthbound == "" {
		nc.log.Error("Tried to send southbound observation without configured subject")
		return common.ErrNotCompleted
	}

	// TODO really re-use connection to KV?
	outMsg := []byte(msg)
	err := nc.conn.Publish(nc.subjectSouthbound, outMsg)
	if err != nil {
		nc.log.Error("Couldn't publish %d bytes msg on %s", len(outMsg), nc.subjectSouthbound)
		return err
	} else {
		nc.log.Debug("Successful publish on '%s'", nc.subjectSouthbound)
	}

	return nil
}

func (nc *natsClient) RemovePrefix(subject string) string {
	subjectCut, ok := strings.CutPrefix(subject, nc.observationSubjectPrefix)
	if !ok {
		nc.log.Warning("Subject '%s' missing prefix '%s'", subject, nc.observationSubjectPrefix)
	}

	return strings.Trim(subjectCut, common.NATS_DELIM)
}

func (nc *natsClient) initObservationBuckets(buckets []ObservationBucket) error {
	for _, b := range buckets {
		_, ok := common.OBS_MAP[b.Observation]
		if !ok {
			nc.log.Error("Unknown observation '%s' for bucket %s", b.Observation, b.Name)
			return common.ErrBadParam
		}

		if b.Ttl <= 0 {
			nc.log.Error("Observation bucket '%s' configured with a TTL <= 0, this is not allowed", b.Name)
			return common.ErrBadParam
		}

		kv, err := nc.setupBucket(b.Bucket)
		if err != nil {
			nc.log.Error("Could not setup bucket '%s': %s", b.Name, err)
			return err
		}

		nc.okvMap.Lock()
		nc.okvMap.m[b.Observation] = kv
		nc.okvMap.Unlock()
	}

	return nil
}

func (nc *natsClient) setupBucket(b Bucket) (jetstream.KeyValue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var kv jetstream.KeyValue
	var err error

	if b.Create {
		kvCfg := jetstream.KeyValueConfig{
			Bucket:         b.Name,
			LimitMarkerTTL: time.Duration(0) * time.Second,
		}

		if b.Ttl <= 0 {
			nc.log.Info("No TTL set for bucket '%s', values will not expire", b.Name)
		} else {
			kvCfg.TTL = time.Duration(b.Ttl) * time.Second
			kvCfg.Description = fmt.Sprintf("TTL: %d seconds", b.Ttl)
		}

		kv, err = nc.js.CreateKeyValue(ctx, kvCfg)

		if errors.Is(err, jetstream.ErrBucketExists) { // TODO config parameter for disabling this behavior?
			nc.log.Warning("A bucket called '%s' already exists in NATS, but with different config. Will attempt to re-use.", b.Name)
			kv, err = nc.attemptReuseBucket(ctx, b.Name)
		}

		if err != nil {
			nc.log.Error("Could not create key value store in NATS: %s", err)
			return nil, err
		}
	} else {
		kv, err = nc.js.KeyValue(ctx, b.Name)
		if err != nil {
			nc.log.Error("Could not find existing bucket '%s': %s", b.Name, err)
			return nil, err
		}
	}

	return kv, nil
}

func (nc *natsClient) attemptReuseBucket(ctx context.Context, bucket string) (jetstream.KeyValue, error) {
	kv, err := nc.js.KeyValue(ctx, bucket)
	if err != nil {
		nc.log.Error("Couldn't find existing bucket '%s'", bucket)
		return nil, err
	}

	status, err := kv.Status(ctx)
	if err != nil {
		nc.log.Error("Couldn't get status of existing bucket '%s': %s", bucket, err)
		return nil, err
	}

	ttl := status.TTL() / time.Second
	lmTtl := status.LimitMarkerTTL() / time.Second
	size := float32(status.Bytes()) / 1000000 /* In megabytes */
	nc.log.Info("Reusing bucket '%s'. TTL: %d, LimitMarketTTL: %d, Size: %.2f MB.", status.Bucket(), ttl, lmTtl, size)

	return kv, nil
}

func (nc *natsClient) genObservationSubject(domain string, obs string) string {
	prefix := strings.Join([]string{nc.observationSubjectPrefix, obs}, common.NATS_DELIM)

	subject := _getSubjectFromFqdn(domain, prefix, "")

	return common.NormalizeNatsSubject(subject)
}

/* If fqdn is "www.example.com", output will be "prefix.com.example.www.suffix" */
func _getSubjectFromFqdn(fqdn, natsPrefix, natsSuffix string) string {
	rev := libtapir.FlipDomainName(fqdn)

	if natsPrefix != "" {
		rev = strings.Trim(natsPrefix, common.NATS_DELIM) + common.NATS_DELIM + rev
	}

	if natsSuffix != "" {
		rev = rev + common.NATS_DELIM + strings.Trim(natsSuffix, common.NATS_DELIM)
	}

	return rev
}
