package telemetry

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/pasztorpisti/go-crc"
)

const (
	// LowLevelProtocol is the name of the real-time telemetry low level protocol.
	LowLevelProtocol = "XtraLib.Stream"
	// LowLevelProtocolVersion is the supported version of the real-time telemetry low level protocol.
	LowLevelProtocolVersion = "0"
)

const (
	// HighLevelProtocol is the name of the real-time telemetry high level protocol.
	HighLevelProtocol = "Tacview.RealTimeTelemetry"
	// HighLevelProtocolVersion is the supported version of the real-time telemetry high level protocol.
	HighLevelProtocolVersion = "0"
)

// HostHandshake is the handshake packet sent by the host.
type HostHandshake struct {
	// LowLevelProtocolVersion is the version of the low level protocol.
	LowLevelProtocolVersion string
	// HighLevelProtocolVersion is the version of the high level protocol.
	HighLevelProtocolVersion string
	// Hostname of the server.
	Hostname string
}

// Encode the host handshake as a string.
func (h *HostHandshake) Encode() (packet string) {
	packet += fmt.Sprintf("%s.%s\n", LowLevelProtocol, LowLevelProtocolVersion)
	packet += fmt.Sprintf("%s.%s\n", HighLevelProtocol, HighLevelProtocolVersion)
	packet += h.Hostname + "\n"
	packet += string(rune(0))
	return
}

// DecodeHostHandshake decodes a host handshake from the given string.
func DecodeHostHandshake(packet string) (HostHandshake, error) {
	handshake := HostHandshake{}
	for _, line := range strings.Split(packet, "\n") {
		if line == "" || line == string(rune(0)) {
			continue
		} else if strings.HasPrefix(line, LowLevelProtocol) {
			handshake.LowLevelProtocolVersion = strings.SplitAfter(line, LowLevelProtocol+".")[1]
		} else if strings.HasPrefix(line, HighLevelProtocol) {
			handshake.HighLevelProtocolVersion = strings.SplitAfter(line, HighLevelProtocol+".")[1]
		} else if len(line) > 0 {
			handshake.Hostname = line
		}
	}
	return handshake, nil
}

// ClientHandshake is the handshake packet sent by the client in response to the host handshake.
type ClientHandshake struct {
	// LowLevelProtocolVersion is the version of the low level protocol.
	LowLevelProtocolVersion string
	// HighLevelProtocolVersion is the version of the high level protocol.
	HighLevelProtocolVersion string
	// Hostname of the client.
	Hostname string
	// PasswordHash is the CRC64 hash of the password.
	PasswordHash string
}

// NewClientHandshake creates a new client handshake using the given client hostname and password.
func NewClientHandshake(hostname string, password string) (handshake *ClientHandshake) {
	return &ClientHandshake{
		LowLevelProtocolVersion:  LowLevelProtocolVersion,
		HighLevelProtocolVersion: HighLevelProtocolVersion,
		Hostname:                 hostname,
		PasswordHash:             hashPassword(password),
	}
}

func hashPassword(password string) string {
	// Convert the password to UTF-16 encoding
	utf16CodeUnits := utf16.Encode([]rune(password))
	buf := make([]byte, len(utf16CodeUnits)*2)
	for i, codeUnit := range utf16CodeUnits {
		binary.LittleEndian.PutUint16(buf[i*2:], codeUnit)
	}

	hash := crc.CRC64WE.Calc(buf)
	return strconv.FormatUint(hash, 16)
}

// Encode the client handshake as a string.
func (h *ClientHandshake) Encode() (packet string) {
	packet += fmt.Sprintf("%s.%s\n", LowLevelProtocol, LowLevelProtocolVersion)
	packet += fmt.Sprintf("%s.%s\n", HighLevelProtocol, HighLevelProtocolVersion)
	packet += h.Hostname + "\n"
	packet += h.PasswordHash + string(rune(0))
	return
}

// DecodeClientHandshake decodes a client handshake from the given string.
func DecodeClientHandshake(packet string) (*ClientHandshake, error) {
	lines := strings.Split(packet, "\n")
	if len(lines) < 4 {
		return nil, errors.New("insufficient lines in handshake packet")
	}
	handshake := &ClientHandshake{}
	if !strings.HasPrefix(lines[0], LowLevelProtocol+".") {
		return nil, errors.New("unexpected low level protocol version")
	}
	handshake.LowLevelProtocolVersion = strings.Split(lines[0], ".")[1]
	if !strings.HasPrefix(lines[1], HighLevelProtocol+".") {
		return nil, errors.New("unexpected high level protocol version")
	}
	handshake.HighLevelProtocolVersion = strings.Split(lines[1], ".")[1]

	if !strings.HasPrefix(lines[2], "Client ") {
		return nil, errors.New("unexpected client hostname")
	}
	handshake.Hostname = strings.Split(lines[2], " ")[1]

	hash, _, ok := strings.Cut(lines[3], string(rune(0)))
	if !ok {
		return nil, errors.New("unable to decode password hash")
	}
	handshake.PasswordHash = hash
	return handshake, nil
}
