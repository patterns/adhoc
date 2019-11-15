package proto

import (
	"bytes"
	"encoding/binary"

	"math"
	"testing"
	"time"
)

func TestHeader(t *testing.T) {
	const (
		expmagic   = "MPS7"
		expversion = 5
		explength  = 123
	)

	// Simulate the Header spec which is 4|1|4 bytes
	buf := []byte{'M', 'P', 'S', '7', expversion, 0, 0, 0, explength}

	// Make io.Reader for the buffer
	nb := bytes.NewBuffer(buf)

	// Start the helper with the input stream
	hlp := NewHelper(nb)

	// Direct call to the Header matching
	hd, err := hlp.header()
	if err != nil {
		t.Error("Failed header extraction")
	}

	if hd.Len() != explength {
		t.Errorf("Expected Header length %d, got %d", explength, hd.Len())
	}
	if hd.Version() != expversion {
		t.Errorf("Expected Header version %v, got %d", expversion, hd.Version())
	}
	if hd.Prefix() != expmagic {
		t.Errorf("Expected Header prefix %s, got %s", expmagic, hd.Prefix())
	}
}

func TestHeaderCompatible(t *testing.T) {
	// Expected field values for the test
	const (
		expmagic   = "MPS7"
		expversion = 5
		explength  = 123
	)

	// Simulate the Header spec which is 4|1|4 bytes
	buf := []byte{'M', 'P', 'S', '7', expversion, 0, 0, 0, explength}

	// Make io.Reader for the buffer
	nb := bytes.NewBuffer(buf)

	// Start the helper with the input stream
	hlp := NewHelper(nb)

	// Run the Header format matching
	// and check the spec version
	vok := hlp.Compatible(byte(expversion))

	if !vok {
		t.Errorf("Expected Compatible version %d", expversion)
	}

	if hlp.Compatible(byte(0)) {
		t.Error("Expected Compatible version mismatch")
	}

	if hlp.state != Ready {
		t.Error("Expected Helper state to have advanced to Ready")
	}
}

func TestHeaderRecordLen(t *testing.T) {
	// Expected field values for the test
	const (
		expmagic   = "MPS7"
		expversion = 5
		explength  = 123
	)

	// Simulate the Header spec which is 4|1|4 bytes
	buf := []byte{'M', 'P', 'S', '7', expversion, 0, 0, 0, explength}

	// Make io.Reader for the buffer
	nb := bytes.NewBuffer(buf)

	// Start the Helper with the input stream
	hlp := NewHelper(nb)

	// Run the Header format matching
	// and check the spec version
	hlp.Compatible(byte(expversion))

	if hlp.Len() != explength {
		t.Errorf("Expected Header length %d, got %d", explength, hlp.Len())
	}

	if hlp.Len() == 0 {
		t.Error("Expected Header length to be non-zero")
	}

	if hlp.state != Ready {
		t.Error("Expected Helper state to have advanced to Ready")
	}
}

func TestRecordFields(t *testing.T) {
	exptype := EndAutopay
	hlp, expuser, expstamp := newRecordTestData(exptype)

	// Run the Record format matching
	rec, err := hlp.Next()
	if err != nil {
		t.Error("Helper Next call failed on error")
	}

	// Verify stamp field
	stamp := uint32(expstamp.Unix())
	if stamp != rec.stamp {
		t.Errorf("Expected Record stamp %x, got %x\n", stamp, rec.stamp)
	}

	// Verify type field
	if exptype != rec.Rectype {
		t.Errorf("Expected Record type %x, got %x\n", exptype, rec.Rectype)
	}

	// Verify user field
	if expuser != rec.User {
		t.Errorf("Expected Record user %d, got %d\n", expuser, rec.User)
	}
}

func TestRecordDebit(t *testing.T) {
	exptype := Debit
	expdollars := 4.99
	hlp, expuser, expstamp := newRecordTestData(exptype, expdollars)

	// Run the Record format matching
	rec, err := hlp.Next()
	if err != nil {
		t.Error("Helper Next call failed on error")
	}

	// Verify stamp field
	stamp := uint32(expstamp.Unix())
	if stamp != rec.stamp {
		t.Errorf("Expected Record stamp %x, got %x\n", stamp, rec.stamp)
	}

	// Verify type field
	if exptype != rec.Rectype {
		t.Errorf("Expected Record type %x, got %x\n", exptype, rec.Rectype)
	}

	// Verify user field
	if expuser != rec.User {
		t.Errorf("Expected Record user %d, got %d\n", expuser, rec.User)
	}

	// Verify dollars field
	if expdollars != rec.Dollars {
		t.Errorf("Expected Record dollars %f, got %f\n", expdollars, rec.Dollars)
	}
}

func TestRecordCredit(t *testing.T) {
	exptype := Credit
	expdollars := 8.99
	hlp, expuser, expstamp := newRecordTestData(exptype, expdollars)

	// Run the Record format matching
	rec, err := hlp.Next()
	if err != nil {
		t.Error("Helper Next call failed on error")
	}

	// Verify stamp field
	stamp := uint32(expstamp.Unix())
	if stamp != rec.stamp {
		t.Errorf("Expected Record stamp %x, got %x\n", stamp, rec.stamp)
	}

	// Verify type field
	if exptype != rec.Rectype {
		t.Errorf("Expected Record type %x, got %x\n", exptype, rec.Rectype)
	}

	// Verify user field
	if expuser != rec.User {
		t.Errorf("Expected Record user %d, got %d\n", expuser, rec.User)
	}

	// Verify dollars field
	if expdollars != rec.Dollars {
		t.Errorf("Expected Record dollars %f, got %f\n", expdollars, rec.Dollars)
	}
}

////////
// test helper routines
////////

func newRecordTestData(rt Rtype, dollars ...float64) (h Helper, u uint64, s time.Time) {
	// Expected field values for the test
	exptype := rt
	expuser := uint64(12)
	expstamp := time.Date(2001, time.September, 9, 1, 46, 40, 0, time.UTC)

	// Simulate the Record spec which is 1|4|8 bytes
	// 1|4|8|8 bytes when type is Debit/Credit
	buf := []byte{
		byte(exptype),
		0x3b, 0x9a, 0xca, 0x00,
		0, 0, 0, 0, 0, 0, 0, 12,
	}

	// For Debit or Credit records, allow dollars field
	if rt == Debit || rt == Credit {
		dol := make([]byte, 8)
		binary.BigEndian.PutUint64(dol, math.Float64bits(dollars[0]))
		buf = append(buf, dol...)
	}

	// Make a io.Reader for buffer
	nb := bytes.NewBuffer(buf)

	// Start the helper with the input stream
	h = NewHelper(nb)
	u = expuser
	s = expstamp

	// Skip Header hlpsing and jump to
	// matching Records
	h.state = Ready

	return
}
