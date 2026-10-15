package classifiers

import (
	"clonopher/internal"
	"encoding/json"
)

type ObjectInfo internal.Envelope

func (i *ObjectInfo) Render() string {
	if i == nil {
		return ""
	}
	b, err := json.Marshal(i)
	if err != nil {
		return ""
	}
	return string(b)
}
