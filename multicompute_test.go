package siavanity

import (
	"fmt"
	"testing"
	"time"
)

func TestMulti(t *testing.T) {
	tSearch := MultiThread{8}
	start := time.Now()
	tSearch.find([16][32]byte{}, "test")
	fmt.Println(time.Since(start))
}
