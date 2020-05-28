package siavanity

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"lukechampine.com/us/merkle"
	"math"
	"os"
	"runtime"
	"testing"
)

func tTestFunc(t *testing.T) {

	fileName := "blake2b.cu"

	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	fileData = prependHeader(os.Args[1], fileData)
	if len(fileData) > sectorSize-leafSize {
		log.Fatal("file is too large")
	}
	// construct intermediate Stack
	var base Stack
	b := bytes.NewBuffer(fileData)
	for base.used != (sectorSize/leafSize)-1 {
		base.appendLeaf(b.Next(leafSize))
	}
	prefix := "eeee"
	prefixLen := len(prefix)
	difficulty := math.Pow(64, float64(prefixLen))
	var roots [16][32]byte
	copy(roots[:], base.roots[:])
	//C.pass_struct((*C.roots)(unsafe.Pointer(&roots)))
	fmt.Printf("Expected difficulty: %.0f\n", difficulty)
	finder := MultiThread{runtime.NumCPU()}
	leaf := finder.find(roots, prefix)
	// double-check
	var sector [sectorSize]byte
	copy(sector[:], fileData)
	copy(sector[len(sector)-leafSize:], leaf[:])
	root := merkle.SectorRoot(&sector)

	fmt.Println("\nFinished!")
	fmt.Printf("Produces Merkle root %s", base64.RawURLEncoding.EncodeToString(root[:]))

	dst := os.Args[1] + ".rawsector"
	if err := ioutil.WriteFile(dst, sector[:], 0666); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nWrote raw sector to", dst)
}
