package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	adhoc "github.com/patterns/adhoc/proto"
)

func main() {
	infile := flag.String("infile", "txnlog.dat", "Path to binary (log) file")
	flag.Parse()

	f, err := os.Open(*infile)
	if err != nil {
		fmt.Printf("Error file open - %s\n", infile)
		return
	}
	defer f.Close()

	r := bufio.NewReader(f)

	parser := adhoc.NewParser(r)

	// Check that file version is compatible (with 1)
	version := byte(1)
	vok := parser.Compatible(version)
	if !vok {
		fmt.Printf("Incorrect MPS7 version - expecting %d\n", version)
		return
	}

	m := loadMap(parser)

	fmt.Println()
	uid := uint64(2456938384156277127)
	fmt.Printf("balance for user %d=%.2f \n", uid, m[uid])
}

func loadMap(parser adhoc.Mps7) map[uint64]float64 {
	var rec adhoc.Record
	var err error
	m := map[uint64]float64{}

	for i := uint32(0); i < parser.Len(); i++ {
		rec, err = parser.Next()
		if err != nil {
			panic(err)
		}

		fmt.Printf("%d) %s\n", i, rec)

		switch rec.Rectype {
		case adhoc.Debit, adhoc.Credit:
			sum := rec.Dollars
			dol, ok := m[rec.User]
			if ok {
				sum = sum + dol
			}
			m[rec.User] = sum
		}
	}

	return m
}
