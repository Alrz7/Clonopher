package helpers

import (
	"fmt"
	"reflect"
	"strings"
)

type InfoOption struct {
	Info    string
	Options []any
}

type InfoOptions []*InfoOption

func (i InfoOptions) ConcatIfExists() string {
	var res string
infoLoop:
	for _, InfoOption := range i {

		for _, option := range InfoOption.Options {
			if option == nil || reflect.ValueOf(option).IsZero() {
				continue infoLoop
			}
		}
		formatted := fmt.Sprintf(InfoOption.Info, InfoOption.Options...)
		res = strings.Join([]string{res, formatted}, "\n")
	}
	return res
}
