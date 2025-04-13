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

type HashAlgorithm string

const (
	// CRC64WE is the CRC-64/WE hash algorithm.
	CRC64WE HashAlgorithm = "CRC-64/WE"
	// CRC32ISOHDLC is the CRC-32/ISO-HDLC hash algorithm.
	CRC32ISOHDLC HashAlgorithm = "CRC-32/ISO-HDLC"
)

func (a HashAlgorithm) String() string {
	return string(a)
}

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
	for line := range strings.SplitSeq(packet, "\n") {
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

	password         string
	hashCRC64WE      *string
	hashCRC32ISOHDLC *string
}

// NewClientHandshake creates a new client handshake using the given client hostname and password.
func NewClientHandshake(hostname string, password string) (handshake *ClientHandshake) {
	return &ClientHandshake{
		LowLevelProtocolVersion:  LowLevelProtocolVersion,
		HighLevelProtocolVersion: HighLevelProtocolVersion,
		Hostname:                 hostname,
		password:                 password,
	}
}

func hashPassword(password string, algorithm HashAlgorithm) string {
	// Convert the password to UTF-16 encoding
	utf16CodeUnits := utf16.Encode([]rune(password))
	buf := make([]byte, len(utf16CodeUnits)*2)
	for i, codeUnit := range utf16CodeUnits {
		binary.LittleEndian.PutUint16(buf[i*2:], codeUnit)
	}

	var hash uint64
	switch algorithm {
	case CRC64WE:
		hash = crc.CRC64WE.Calc(buf)
	case CRC32ISOHDLC:
		hash = uint64(crc.CRC32ISOHDLC.Calc(buf))
	}
	return strconv.FormatUint(hash, 16)
}

// Encode the client handshake as a string.
func (h *ClientHandshake) Encode(algorithm HashAlgorithm) (packet string) {
	packet += fmt.Sprintf("%s.%s\n", LowLevelProtocol, LowLevelProtocolVersion)
	packet += fmt.Sprintf("%s.%s\n", HighLevelProtocol, HighLevelProtocolVersion)
	packet += h.Hostname + "\n"
	packet += hashPassword(h.password, algorithm)
	packet += string(rune(0))
	return
}

// HashCRC64WE returns the CRC64WE hash of the password. If the handshake was created using [NewClientHandshake],
// the hash is computed from the provided password. If the handshake was created using [DecodeClientHandshake],
// the hash is read from the decoded packet, or is "0" if no CRC64WE hash was found.
func (h *ClientHandshake) HashCRC64WE() string {
	if h.hashCRC64WE != nil {
		return *h.hashCRC64WE
	}
	return hashPassword(h.password, CRC64WE)
}

// Returns the CRC32ISOHDLC hash of the password. If the handshake was created using [NewClientHandshake],
// the hash is computed from the provided password. If the handshakre was created using [DecodeClientHandshake],
// the hash is read from the decoded packet, or is "0" if no CRC32ISOHDLC hash was found.
func (h *ClientHandshake) HashCRC32ISOHDLC() string {
	if h.hashCRC32ISOHDLC != nil {
		return *h.hashCRC32ISOHDLC
	}
	return hashPassword(h.password, CRC32ISOHDLC)
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
	if len(hash) == 8 {
		handshake.hashCRC32ISOHDLC = &hash
	} else if len(hash) == 16 {
		handshake.hashCRC64WE = &hash
	} else {
		return nil, errors.New("unexpected password hash length")
	}
	return handshake, nil
}
