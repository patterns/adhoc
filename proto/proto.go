package proto

import (
	"github.com/pkg/errors"
	"io"
)

// Helper is the file helper
type Helper struct {
	hdr   Header
	state pstate
	dec   *Decoder
}

// pstate are the different helper stages
type pstate uint32

const (
	None pstate = 1 << iota
	Starting
	Compatible
	Ready
	Recovery
)

// ProtData is the data type supported in Decode
type ProtData interface {
	Protocol()
}

// NewHelper makes a new helper given the input stream
func NewHelper(r io.Reader) Helper {
	return Helper{
		state: Starting,
		dec:   NewDecoder(r),
	}
}

// Next extracts another record
func (p Helper) Next() (Record, error) {
	if p.state != Ready {
		err := errors.New("Not Ready - compatible version required")
		return Record{}, err
	}

	var rec Record
	err := p.dec.Decode(&rec)
	if err != nil {
		err = errors.Wrap(err, "Next failed - decode error")
		return Record{}, err
	}

	return rec, nil
}

// header extracts header fields
func (p Helper) header() (Header, error) {
	if p.state != Starting {
		err := errors.New("Wrong state - header already consumed")
		return Header{}, err
	}

	var hdr Header
	err := p.dec.Decode(&hdr)
	if err != nil {
		err = errors.Wrap(err, "header failed - Decode error")
		return Header{}, err
	}

	return hdr, nil
}

// Compatible checks version is supported
func (p *Helper) Compatible(ver byte) bool {
	// todo ?Do we accept future/greater version values

	// Re-entrant, only process header fields once
	if p.state == Starting {
		var err error
		p.hdr, err = p.header()
		if err != nil {
			// Mark helper state as in recovery,
			// to do self-repair
			p.state = Recovery
			return false
		}
		p.state = Compatible
	}

	if p.hdr.version == ver && p.hdr.prefix == prefixMPS7 {
		p.state = Ready
		return true
	}
	return false
}

// Len is the record count
func (p Helper) Len() uint32 {
	if p.state&(Ready|Compatible) != 0 {
		// Ready or Compatible state are when
		// the header fields are okay
		return p.hdr.Len()
	}

	// todo ?Is this an (fatal) error or
	return 0
}
