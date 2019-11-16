package proto

import (
	"fmt"
)

type Record struct {
	Rectype Rtype
	stamp   uint32
	User    uint64

	// todo ?eliminate because we embed the common fields
	Dollars float64
}

type Rtype byte

const (
	Debit Rtype = iota
	Credit
	StartAutopay
	EndAutopay
)

// Protocol is required for ProtData interface
func (r Record) Protocol() {
	//todo indicate MPS7 support
}

func (r Record) String() string {
	return fmt.Sprintf("%s, %d, %d, %f",
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
