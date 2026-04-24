package state

import "encoding/json"

func jsonMarshalIndent(rec Record) ([]byte, error) {
	return json.MarshalIndent(rec, "", "  ")
}
