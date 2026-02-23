// Package brevity contains types and models for air combat communication brevity.
package brevity

// LastCaller is a placeholder callsign used when the actual callsign is unknown.
const LastCaller = "Last caller"

// Response is implemented by all brevity response and call types that a GCI controller can send.
// The unexported method prevents types outside this package from satisfying the interface.
type Response interface {
	isBrevityResponse()
}
