package types

import "github.com/lithammer/shortuuid/v3"

// GUID is a unique identifier for an SRS network client. Each client generates a 22-byte GUID on startup. GUIDs are encoded in base57.
type GUID string

// GUIDLength is the length of a GUID in bytes.
const GUIDLength = 22

// NewGUID generates a new GUID of length GUIDLength.
func NewGUID() (guid GUID) {
	for len([]byte(guid)) != GUIDLength {
		guid = GUID(shortuuid.New())
	}
	return
}
