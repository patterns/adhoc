package proto

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"io"
	"math"
)

// prefixMPS7 is the required header prefix for protocol
const prefixMPS7 = "MPS7"

// Decoder decodes records from a MPS7 stream
type Decoder struct {
	r io.Reader
}

// NewDecoder reads from the MPS7 stream
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (dec *Decoder) Decode(v ProtData) error {

	switch test := v.(type) {
	case *Record:
		return dec.decodeRec(test)

	case *Header:
		return dec.decodeHdr(test)

	default:
		return errors.New("Decode failed - unknown type")
	}

	return nil
}

// Decode reads the next MPS7 encoded record and stores it in rec
func (dec *Decoder) decodeRec(rec *Record) error {

	// MPS7 fixed-size spec
	var data struct {
		Type  byte
		Stamp [4]byte
		User  [8]byte
	}

	// Read stream in network byte order
	err := binary.Read(dec.r, binary.BigEndian, &data)
	if err != nil {
		return errors.Wrap(err, "Decode failed - data does not match 1|4|8 bytes")
	}

	rtype := Rtype(data.Type)
	stamp := binary.BigEndian.Uint32(data.Stamp[:])
	user := binary.BigEndian.Uint64(data.User[:])

	*rec = Record{
		Rectype: rtype,
		stamp:   stamp,
		User:    user,
	}

	// Debit/credit means there is a 8 byte field for dollars
	if rtype == Debit || rtype == Credit {
		buf := make([]byte, 8)
		err = binary.Read(dec.r, binary.BigEndian, &buf)
		if err != nil {
			return errors.Wrap(err, "Decode failed - dollars not read")
		}

		bits := binary.BigEndian.Uint64(buf[:])
		rec.Dollars = math.Float64frombits(bits)
	}

	return nil
}

// Decode reads the MPS7 encoded header and stores it in hdr
func (dec *Decoder) decodeHdr(hdr *Header) error {
	// MPS7 fixed-size spec
	var data struct {
		Magic   [4]byte
		Version byte
		Length  [4]byte
	}

	// Read in network byte order
	err := binary.Read(dec.r, binary.BigEndian, &data)
	if err != nil {
		return errors.Wrap(err, "Decode failed - header binary.Read")
		return errors.Wrap(err, "Decode failed - data does not match 4|1|4 bytes")
	}

	pre := string(data.Magic[:])
	ver := data.Version
	len := binary.BigEndian.Uint32(data.Length[:])

	*hdr = Header{prefix: pre, version: ver, length: len}
	return nil
}
