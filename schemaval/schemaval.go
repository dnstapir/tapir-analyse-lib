package schemaval

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/dnstapir/tapir-analyse-lib/common"
	"github.com/dnstapir/tapir-analyse-lib/logger"
)

const c_ID = "tal-validator"

type Conf struct {
	Debug            bool   `toml:"debug"`
	SchemaDir        string `toml:"schema_dir"`
	AllowNoSchema    bool   `toml:"allow_no_schema"`
	AllowNoVerKeys   bool   `toml:"allow_no_verification_keys"`
	VerificationKeys string `toml:"verification_keys"`
	SigningKey       string `toml:"signing_key"`
	Log              common.Logger
}

type schemaval struct {
	id         string
	log        common.Logger
	schemas    map[string]*jsonschema.Schema
	verkeys    jwk.Set
	signingKey jwk.Key
}

func Create(conf Conf) (*schemaval, error) {
	s := new(schemaval)
	s.id = c_ID

	if conf.Log == nil {
		log := logger.New(
			logger.Conf{
				Debug: conf.Debug,
			})
		s.log = log
	} else {
		s.log = conf.Log
	}
	s.log.Debug("%s: debug logging enabled", s.id)

	if conf.VerificationKeys == "" {
		if conf.AllowNoVerKeys {
			s.log.Warning("No verification keys specified, will not check signatures")
			s.verkeys = nil
		} else {
			return nil, common.ErrBadParam
		}
	} else {
		keys, err := jwk.ReadFile(conf.VerificationKeys)
		if err != nil {
			s.log.Error("Couldn't read verification keys file: %s", err)
			return nil, errors.New("bad verification keys file")
		}
		s.verkeys = keys
		s.log.Info("Read %d verification keys", s.verkeys.Len())
	}

	if conf.SchemaDir == "" {
		if conf.AllowNoSchema {
			s.log.Warning("No schemadir specified, will accept anything")
			s.schemas = nil
		} else {
			return nil, common.ErrBadParam
		}
	} else {
		files, err := os.ReadDir(conf.SchemaDir)
		if err != nil {
			s.log.Error("Error reading schema dir %s", conf.SchemaDir)
			return nil, err
		}
		if len(files) == 0 {
			s.log.Error("No schemas found in %s", conf.SchemaDir)
			return nil, errors.New("no schemas found")
		}

		s.schemas = make(map[string]*jsonschema.Schema)
		c := jsonschema.NewCompiler()

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			fullName := filepath.Join(conf.SchemaDir, file.Name())
			schema, err := c.Compile(fullName)
			if err != nil {
				s.log.Error("Compiling schema %s failed: %s", file.Name(), err)
				return nil, err
			}

			s.schemas[schema.ID] = schema
		}
	}

	if conf.SigningKey == "" {
		s.log.Warning("No signing key specified, will not be able to sign messages")
	} else {
		keyFile, err := os.ReadFile(conf.SigningKey)
		if err != nil {
			s.log.Error("Could not read signing key file, err: '%s'", err)
			return nil, err
		}

		keyParsed, err := jwk.ParseKey(keyFile)
		if err != nil {
			s.log.Error("Could not parse signing key file, err: '%s'", err)
			return nil, err
		}

		isPrivate, err := jwk.IsPrivateKey(keyParsed)
		if err != nil {
			s.log.Error("Could not check if key is private, err: '%s'", err)
			return nil, err
		}

		if !isPrivate {
			s.log.Error("Signing key file '%s' is not private", conf.SigningKey)
			return nil, errors.New("signing key must be private")
		}

		_, hasAlg := keyParsed.Algorithm()
		if !hasAlg {
			s.log.Error("Signing key missing an \"alg\" field")
			return nil, common.ErrBadJWK
		}

		s.signingKey = keyParsed
	}

	return s, nil
}

func (s *schemaval) ValidateWithID(data []byte, id string) bool {
	if s.schemas == nil {
		s.log.Warning("Proper validation of %d bytes skipped", len(data))
		return true
	}

	schema, ok := s.schemas[id]
	if !ok {
		s.log.Warning("Requested schema %s not found", id)
		return false
	}

	dataReader := bytes.NewReader(data)
	obj, err := jsonschema.UnmarshalJSON(dataReader)
	if err != nil {
		s.log.Warning("Error unmarshalling byte stream into JSON object: %s", err)
		return false
	}

	err = schema.Validate(obj)
	if err != nil {
		s.log.Debug("Validation error '%s'", err)
		return false
	}

	return true
}

func (s *schemaval) VerifySignature(sig []byte) ([]byte, error) {
	if s.verkeys == nil {
		s.log.Error("Skipping signature verification of %d bytes", len(sig))
		return nil, common.ErrNotCompleted
	}

	data, err := jws.Verify(sig, jws.WithKeySet(s.verkeys))
	if err != nil {
		s.log.Error("Could not verify data: %s", err)
		return nil, err
	}

	return data, nil
}

func (s *schemaval) SignData(data []byte) ([]byte, error) {
	if s.signingKey == nil {
		s.log.Error("No signing key configured.")
		return nil, common.ErrNotCompleted
	}

	alg, _ := s.signingKey.Algorithm()
	signedData, err := jws.Sign(data, jws.WithJSON(), jws.WithKey(alg, s.signingKey))
	if err != nil {
		s.log.Error("Couldn't sign data: %s", err)
		return nil, err
	}

	return signedData, nil
}
