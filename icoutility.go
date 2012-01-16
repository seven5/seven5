package seven5

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

//DumpIcoFileAsBase64 prints to stdout the base64 encoding of the file pointed to by file
//parameter.  This can then be used with the FavicoGuise.SetFavIco() to set the icon for the 
//web application.  Usually 16x16 or 32x32 icons are best and must be "ico" (windows icon)
//format.
func DumpIcoFileAsBase64(filename string) {
	buffer := new(bytes.Buffer)
	enc := base64.NewEncoder(base64.StdEncoding, buffer)
	file, e := os.Open(filename)
	if e != nil {
		fmt.Fprintf(os.Stderr, "open:%s\n", e.Error())
		return
	}
	in := make([]byte, 10000) //slightly bigger than file contents
	n, e := io.ReadFull(file, in)
	if n != 9854 && e != nil {
		fmt.Fprintf(os.Stderr, "read full:%s\n", e.Error())
		return
	}
	if n != 9854 {
		fmt.Fprintf(os.Stderr, "read %d bytes but expected 9854\n", n)
	}
	n, e = enc.Write(in[0:n])
	if e != nil {
		fmt.Fprintf(os.Stderr, "write/encode:%s\n", e.Error())
		return
	}
	if n != 9854 {
		fmt.Fprintf(os.Stderr, "encoded %d bytes but expected 9854\n", n)
	}
	enc.Close()

	size := len(buffer.Bytes())

	incr := 64
	for i := 0; i < size; i += incr {
		if i+incr >= size {
			fmt.Fprintf(os.Stdout, "%s\n", buffer.Bytes()[i:size])
		} else {
			fmt.Fprintf(os.Stdout, "%s\n", buffer.Bytes()[i:i+incr])
		}
	}
	return
}
