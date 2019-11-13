package adhoc

import (
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"math"
)

// Mps7 is the Mps7 format parser
type Mps7 struct {
	rdr   io.Reader
	hdr   Header
	state pstate
}

// Header is the print friendly translation of the header fields
type Header struct {
	magic   string
	version byte
	length  uint32
}

type pstate int

const (
	None pstate = iota
	Starting
	Compatible
	Ready
)

type Rtype byte

const (
	Debit Rtype = iota
	Credit
	StartAutopay
	EndAutopay
)

type Record struct {
	Rectype Rtype
	stamp   uint32
	User    uint64
	Dollars float64
}

func NewParser(r io.Reader) Mps7 {
	return Mps7{
		rdr:   r,
		state: Starting,
	}
}

// Magic is the magic field from the Header data
func (h Header) Magic() string {
	return h.magic
}

// Version is the version field from the Header data
func (h Header) Version() byte {
	return h.version
}

// Len is the record count from the Header data
func (h Header) Len() uint32 {
	return h.length
}

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

func (m Mps7) header() Header {

	// header format spec
	var data struct {
		Magic   [4]byte
		Version byte
		Length  [4]byte
	}

	// todo ?Do we want a reset-able reader and rewind to the beginning
	// in case steps are done out-of-order
	err := binary.Read(m.rdr, binary.BigEndian, &data)
	if err != nil {
		log.Fatal("Header binary.Read failed:", err)
	}

	mag := string(data.Magic[:])
	v := data.Version
	l := binary.BigEndian.Uint32(data.Length[:])

	return Header{magic: mag, version: v, length: l}
}

func (m *Mps7) Compatible(ver byte) bool {
	if m.state == Starting {
		m.hdr = m.header()
		m.state = Compatible
	}

	if m.hdr.version == ver && m.hdr.magic == "MPS7" {
		m.state = Ready
		return true
	}
	return false
}

func (m Mps7) Len() uint32 {
	if m.state < Compatible {
		// todo ?Is this an (fatal) error or
		return 0
	}
	return m.hdr.Len()
}


func (r Record) String() string {
	return fmt.Sprintf("%s, %d, %d - %f",
		r.Rectype, r.stamp, r.User, r.Dollars)
}

func (r Rtype) String() string {
	switch r {
	case Debit:
		return "Debit"
	case Credit:
		return "Credit"
	case StartAutopay:
		return "StartAutopay"
	case EndAutopay:
		return "EndAutopay"
	default:
		return "None"
	}
}
