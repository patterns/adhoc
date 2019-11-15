package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	adhoc "github.com/patterns/adhoc/proto"
)

var plog *adhoc.Plog

func main() {
	infile := flag.String("infile", "txnlog.dat", "Path to binary (log) file")
	verbose := flag.Bool("verbose", false, "Enable extra messages")
	flag.Parse()
	var err error
	plog, err = adhoc.NewLog(*verbose)
	if err != nil {
		fmt.Println("Debug logger failed")
		return
	}
	defer plog.Close()

	f, err := os.Open(*infile)
	if err != nil {
		plog.Verbose(fmt.Sprintf("Error file open - %s", *infile))
		// Raise the exception
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)

	helper := adhoc.NewHelper(r)

	// Check that file version is compatible (with 1)
	version := byte(1)
	vok := helper.Compatible(version)
	if !vok {
		plog.Verbose(fmt.Sprintf("Incorrect MPS7 version - expecting %d", version))
		return
	}

	kv, tot := loadMap(helper)

	// Show summary
	plog.Info(fmt.Sprintf("total credit amount=%.2f", tot[adhoc.Credit]))
	plog.Info(fmt.Sprintf("total debit amount=%.2f", tot[adhoc.Debit]))
	plog.Info(fmt.Sprintf("autopays started=%d", int(tot[adhoc.StartAutopay])))
	plog.Info(fmt.Sprintf("autopays ended=%d", int(tot[adhoc.EndAutopay])))
	uid := uint64(2456938384156277127)
	plog.Info(fmt.Sprintf("balance for user %d=%.2f", uid, kv[uid]))
}

// loadMap loops through records and makes calculations
func loadMap(helper adhoc.Helper) (map[uint64]float64, map[adhoc.Rtype]float64) {
	var rec adhoc.Record
	var err error

	// initialize the results maps
	m := map[uint64]float64{}
	tot := map[adhoc.Rtype]float64{}
	tot[adhoc.Credit] = 0
	tot[adhoc.Debit] = 0

	for i := uint32(0); i < helper.Len(); i++ {
		rec, err = helper.Next()
		if err != nil {
			panic(err)
		}

		plog.Verbose(fmt.Sprintf("%d) %s", i, rec))

		switch rec.Rectype {
		case adhoc.StartAutopay, adhoc.EndAutopay:
			if count, ok := tot[rec.Rectype]; ok {
				tot[rec.Rectype] = count + 1
			} else {
				tot[rec.Rectype] = 1
			}

		case adhoc.Debit, adhoc.Credit:
			// Accumulate money of user
			sum := rec.Dollars
			if dol, ok := m[rec.User]; ok {
				sum = sum + dol
			}
			m[rec.User] = sum

			// Also total the money as aggregate
			tot[rec.Rectype] = tot[rec.Rectype] + rec.Dollars
		}
	}

	return m, tot
}
