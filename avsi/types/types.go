package types

import (
	"bytes"
	"encoding/json"

	"github.com/cosmos/gogoproto/jsonpb"
)

const (
	CodeTypeOK uint32 = 0
)

// IsOK returns true if Code is OK.
func (r ResponseQuery) IsOK() bool {
	return r.Code == CodeTypeOK
}

// IsErr returns true if Code is something other than OK.
func (r ResponseQuery) IsErr() bool {
	return r.Code != CodeTypeOK
}

//---------------------------------------------------------------------------
// override JSON marshaling so we emit defaults (ie. disable omitempty)

var (
	jsonpbMarshaller = jsonpb.Marshaler{
		EnumsAsInts:  true,
		EmitDefaults: true,
	}
	jsonpbUnmarshaller = jsonpb.Unmarshaler{}
)

func (r *ResponseQuery) MarshalJSON() ([]byte, error) {
	s, err := jsonpbMarshaller.MarshalToString(r)
	return []byte(s), err
}

func (r *ResponseQuery) UnmarshalJSON(b []byte) error {
	reader := bytes.NewBuffer(b)
	return jsonpbUnmarshaller.Unmarshal(reader, r)
}

// Some compile time assertions to ensure we don't
// have accidental runtime surprises later on.

// jsonEncodingRoundTripper ensures that asserted
// interfaces implement both MarshalJSON and UnmarshalJSON
type jsonRoundTripper interface {
	json.Marshaler
	json.Unmarshaler
}

var _ jsonRoundTripper = (*ResponseQuery)(nil)
