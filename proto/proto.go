package proto

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"io"
	"math"
)

// Mps7 is the Mps7 format parser
type Mps7 struct {
	rdr   io.Reader
	hdr   Header
	state pstate
}

// pstate are the different parser stages
type pstate uint32

const (
	None pstate = 1 << iota
	Starting
	Compatible
	Ready
	Recovery
)

// magictag is the required header prefix for protocol
const magictag = "MPS7"

// NewParser makes a new parser given the input stream
func NewParser(r io.Reader) Mps7 {
	return Mps7{
		rdr:   r,
		state: Starting,
	}
}

// Next extracts the record fields from the stream
func (m Mps7) Next() (r Record, err error) {
	r = Record{}
	err = nil
	if m.state != Ready {
		err = errors.New("Not Ready - must satisfy Compatibility check first")
		return
	}

	// record format spec
	var data struct {
		Type  byte
		Stamp [4]byte
		User  [8]byte
	}

	// Use network byte order for the stream read
	err = binary.Read(m.rdr, binary.BigEndian, &data)
	if err != nil {
		err = errors.Wrap(err, "Record read - cannot match format")
		return
	}

	user := binary.BigEndian.Uint64(data.User[:])
	rectype := Rtype(data.Type)
	stamp := binary.BigEndian.Uint32(data.Stamp[:])
	r = Record{
		Rectype: rectype,
		stamp:   stamp,
		User:    user,
	}

	// debit/credit means there is a 8 byte field for dollars
	if rectype == Debit || rectype == Credit {
		buf := make([]byte, 8)
		err = binary.Read(m.rdr, binary.BigEndian, &buf)
		if err != nil {
			err = errors.Wrap(err, "Record read - cannot match dollars")
			return
		}

		bits := binary.BigEndian.Uint64(buf[:])
		r.Dollars = math.Float64frombits(bits)
	}

	return
}

// header extracts header fields
func (m Mps7) header() (hdr Header, err error) {
	err = nil
	hdr = Header{}
	if m.state != Starting {
		err = errors.New("Wrong state - header already consumed")
		return
	}

	// header format spec
	var data struct {
		Magic   [4]byte
		Version byte
		Length  [4]byte
	}

	// todo ?Do we want a reset-able reader and rewind to the beginning
	// in case steps are done out-of-order
	err = binary.Read(m.rdr, binary.BigEndian, &data)
	if err != nil {
		err = errors.Wrap(err, "Header binary.Read")
		return
	}

	mag := string(data.Magic[:])
	v := data.Version
	l := binary.BigEndian.Uint32(data.Length[:])

	hdr = Header{magic: mag, version: v, length: l}
	return
}

// Compatible checks version is supported
func (m *Mps7) Compatible(ver byte) bool {
	// todo ?Do we accept future/greater version values

	// Re-entrant, only process header fields once
	if m.state == Starting {
		var err error
		m.hdr, err = m.header()
		if err != nil {
			// Mark parser state as in recovery,
			// to do self-repair
			m.state = Recovery
			return false
		}
		m.state = Compatible
	}

	if m.hdr.version == ver && m.hdr.magic == magictag {
		m.state = Ready
		return true
	}
	return false
}

// Len is the record count
func (m Mps7) Len() uint32 {
	if m.state&(Ready|Compatible) != 0 {
		// Ready or Compatible state are when
		// the header fields are okay
		return m.hdr.Len()
	}

	// todo ?Is this an (fatal) error or
	return 0
}
