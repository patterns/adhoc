package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	adhoc "github.com/patterns/adhoc/proto"
)

// protomgr holds the cmd results
type protomgr struct {
	l       *adhoc.Plog
	rdr     io.ReadCloser
	infile  *string
	verbose *bool
	version byte
	qry     *uint64
	agg     map[adhoc.Rtype]float64
	bal     map[uint64]float64
}

func main() {
	mgr := newMgr()
	defer mgr.Close()

	worker := adhoc.NewWorker(mgr.rdr)

	// Check file version is compatible (with '1')
	vok := worker.Compatible(mgr.version)
	if !vok {
		mgr.l.Verbose(fmt.Sprintf("MPS7 version - %x unsupported", mgr.version))
		return
	}

	dispatch(worker, &mgr)
	summary(mgr)
}

func newMgr() protomgr {
	var err error

	// Initialize params
	mgr := protomgr{version: byte(1)}
	mgr.infile = flag.String("infile", "txnlog.dat", "Path to binary (log) file")
	mgr.verbose = flag.Bool("verbose", false, "Enable extra messages")
	mgr.qry = flag.Uint64("qry", 2456938384156277127, "User ID to look-up")
	flag.Parse()

	// Enable logging
	mgr.l, err = adhoc.NewLog(*mgr.verbose)
	if err != nil {
		fmt.Println("Debug logger failed")
		panic(err)
	}

	mgr.rdr, err = os.Open(*mgr.infile)
	if err != nil {
		mgr.l.Verbose(fmt.Sprintf("Error file open - %s", *mgr.infile))
		panic(err)
	}
	return mgr
}

func (m protomgr) Close() {
	m.l.Close()
	m.rdr.Close()
}

// dispatch loops through records and runs associated actions
func dispatch(worker adhoc.Worker, mgr *protomgr) {
	var rec adhoc.Record
	var err error

	// initialize the results maps
	mgr.bal = map[uint64]float64{}
	mgr.agg = map[adhoc.Rtype]float64{}
	mgr.agg[adhoc.Credit] = 0
	mgr.agg[adhoc.Debit] = 0

	for i := uint32(0); i < worker.Len(); i++ {
		rec, err = worker.Next()
		if err != nil {
			panic(err)
		}

		mgr.l.Verbose(rec.String())

		switch rec.Rectype {
		case adhoc.StartAutopay, adhoc.EndAutopay:
			actionAggregate(rec, mgr)

		case adhoc.Debit, adhoc.Credit:
			actionBalance(rec, mgr)
		}
	}
}

// actionAggregate are steps for Start/End autopay transactions
func actionAggregate(rec adhoc.Record, mgr *protomgr) {
	if count, ok := mgr.agg[rec.Rectype]; ok {
		mgr.agg[rec.Rectype] = count + 1
	} else {
		mgr.agg[rec.Rectype] = 1
	}
}

// actionBalance are steps for Debit/Credit transactions
func actionBalance(rec adhoc.Record, mgr *protomgr) {
	// Accumulate money of user
	sum := rec.Dollars
	if rec.Rectype == adhoc.Debit {
		// debit subtracts from balance
		sum = (-1) * rec.Dollars
	}
	if dol, ok := mgr.bal[rec.User]; ok {
		sum = sum + dol
	}
	mgr.bal[rec.User] = sum

	// Also total the money as aggregate
	mgr.agg[rec.Rectype] = mgr.agg[rec.Rectype] + rec.Dollars
}

// summary displays results
func summary(mgr protomgr) {
	mgr.l.Info(fmt.Sprintf("total credit amount=%.2f", mgr.agg[adhoc.Credit]))
	mgr.l.Info(fmt.Sprintf("total debit amount=%.2f", mgr.agg[adhoc.Debit]))
	mgr.l.Info(fmt.Sprintf("autopays started=%d", int(mgr.agg[adhoc.StartAutopay])))
	mgr.l.Info(fmt.Sprintf("autopays ended=%d", int(mgr.agg[adhoc.EndAutopay])))
	mgr.l.Info(fmt.Sprintf("balance for user %d=%.2f", *mgr.qry, mgr.bal[*mgr.qry]))
}
