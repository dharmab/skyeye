package types

import "github.com/lithammer/shortuuid/v3"

type GUID string

const GUIDLength = 22

func NewGUID() (guid GUID) {
	for len([]byte(guid)) != GUIDLength {
		guid = GUID(shortuuid.New())
	}
	return
}
