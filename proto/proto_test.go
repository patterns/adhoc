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
	/*
		var n int
		buf := make([]byte, 0, 9)
		b := bytes.NewBuffer(buf)

		// Put the magic field into the buffer
		n, _ = b.WriteString(expmagic)
		if n != 4 {
			t.Error("Test data error magic")
		}

		// Put the version field into the buffer
		b.WriteByte(expversion)

		// Put the record count field into the buffer
		count := []byte{explength,0,0,0}
		n, _ = b.Write(count)
		if n != 4 {
			t.Error("Test data error count")
		}
	*/

	// Simulate the Header spec which is 4|1|4 bytes
	buf := []byte{'M', 'P', 'S', '7', expversion, 0, 0, 0, explength}

	// Make io.Reader for the buffer
	nb := bytes.NewBuffer(buf)

	// Start the parser with the input stream
	par := NewParser(nb)

	// Direct call to the Header matching
	hd, err := par.header()
	if err != nil {
		t.Error("Failed header extraction")
	}

	if hd.Len() != explength {
		t.Errorf("Expected Header length %d, got %d", explength, hd.Len())
	}
	if hd.Version() != expversion {
		t.Errorf("Expected Header version %v, got %d", expversion, hd.Version())
	}
	if hd.Magic() != expmagic {
		t.Errorf("Expected Header magic %s, got %s", expmagic, hd.Magic())
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

	// Start the parser with the input stream
	par := NewParser(nb)

	// Run the Header format matching
	// and check the spec version
	vok := par.Compatible(byte(expversion))

	if !vok {
		t.Errorf("Expected Compatible version %d", expversion)
	}

	if par.Compatible(byte(0)) {
		t.Error("Expected Compatible version mismatch")
	}

	if par.state != Ready {
		t.Error("Expected Parser state to have advanced to Ready")
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

	// Start the Parser with the input stream
	par := NewParser(nb)

	// Run the Header format matching
	// and check the spec version
	par.Compatible(byte(expversion))

	if par.Len() != explength {
		t.Errorf("Expected Header length %d, got %d", explength, par.Len())
	}

	if par.Len() == 0 {
		t.Error("Expected Header length to be non-zero")
	}

	if par.state != Ready {
		t.Error("Expected Parser state to have advanced to Ready")
	}
}

func TestRecordFields(t *testing.T) {
	exptype := EndAutopay
	par, expuser, expstamp := newRecordTestData(exptype)

	// Skip Header parsing and jump to
	// matching Records
	par.state = Ready

	// Run the Record format matching
	rec, err := par.Next()
	if err != nil {
		t.Error("Parser Next call failed on error")
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
	par, expuser, expstamp := newRecordTestData(exptype, expdollars)

	// Skip Header parsing and jump to
	// matching Records
	par.state = Ready

	// Run the Record format matching
	rec, err := par.Next()
	if err != nil {
		t.Error("Parser Next call failed on error")
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
	par, expuser, expstamp := newRecordTestData(exptype, expdollars)

	// Skip Header parsing and jump to
	// matching Records
	par.state = Ready

	// Run the Record format matching
	rec, err := par.Next()
	if err != nil {
		t.Error("Parser Next call failed on error")
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

func newRecordTestData(rt Rtype, dollars ...float64) (m Mps7, u uint64, s time.Time) {
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
		////0, 0, 'i', 't', 'h', '1', '2', '3',
	}

	// For Debit or Credit records, allow dollars field
	if rt == Debit || rt == Credit {
		dol := make([]byte, 8)
		binary.BigEndian.PutUint64(dol, math.Float64bits(dollars[0]))
		buf = append(buf, dol...)
	}

	// Make a io.Reader for buffer
	nb := bytes.NewBuffer(buf)

	// Start the parser with the input stream
	m = NewParser(nb)
	u = expuser
	s = expstamp
	return
}
