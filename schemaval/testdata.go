package schemaval

import _ "embed"

//go:embed testdata/data/simple.json
var testdataSimple []byte

/* Generated with github.com/go-jose/go-hose v4.1.3 and key a.json */
//go:embed testdata/data/simple_signed.json
var testdataSimpleSigned []byte

//go:embed testdata/data/new_qname_good.json
var testdataNewQnameGood []byte

/* Generated with github.com/go-jose/go-hose v4.1.3 and key a.json */
//go:embed testdata/data/new_qname_good_signed.json
var testdataNewQnameGoodSigned []byte
