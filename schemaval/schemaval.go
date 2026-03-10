package schemaval

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/dnstapir/tapir-analyse-lib/common"
	"github.com/dnstapir/tapir-analyse-lib/logger"
)

const c_ID = "tal-validator"

type Conf struct {
	Debug         bool   `toml:"debug"`
	SchemaDir     string `toml:"schema_dir"`
	AllowNoSchema bool   `toml:"allow_no_schema"`
	Log           common.Logger
}

type schemaval struct {
	id            string
	log           common.Logger
	allowNoSchema bool
	schemas       map[string]*jsonschema.Schema
}

func Create(conf Conf) (*schemaval, error) {
	s := new(schemaval)
	s.id = c_ID
	s.allowNoSchema = conf.AllowNoSchema

	if conf.Log == nil {
		log := logger.New(
			logger.Conf{
				Debug: conf.Debug,
			})
		s.log = log
	} else {
		s.log = conf.Log
	}
	defer s.log.Debug("%s: debug logging enabled", s.id)

	if conf.SchemaDir == "" {
		if s.allowNoSchema {
			s.log.Warning("No schemadir specified, will accept anything")
			return s, nil
		} else {
			return nil, common.ErrBadParam
		}
	}

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
		fullName := filepath.Join(conf.SchemaDir, file.Name())
		schema, err := c.Compile(fullName)
		if err != nil {
			s.log.Error("Compiling schema %s failed: %s", file, err)
			return nil, err
		}

		s.schemas[schema.ID] = schema
	}

	return s, nil
}

func (s *schemaval) ValidateWithID(data []byte, id string) bool {
	if s.schemas == nil && s.allowNoSchema {
		s.log.Debug("Proper validation of %d bytes skipped", len(data))
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
