package main

import (
	"crypto/aes"

	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s key", os.Args[0])
	}
	arg := strings.TrimSpace(os.Args[1])
	if len(arg) != aes.BlockSize {
		log.Fatalf("argument must be %d characters (see https://strongpasswordgenerator.com)", aes.BlockSize)
	}
	buf := make([]byte, 2*aes.BlockSize)
	_ = hex.Encode(buf, []byte(arg))
	fmt.Printf("%s\n", string(buf))
}
