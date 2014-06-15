package seven5

import (
	"fmt"
	"net/http"
)

//generateStringPrinter creates a function suitable for use with a ServeMux's handle func.
func generateStringPrinter(content string, contentType string) func(http.ResponseWriter, *http.Request) {
	return generateBinPrinter([]byte(content), contentType)
}

//generateBinPrinter creates a function suitable for use with a ServeMux's handle func. It writes out the
//content type as a sequence bytes.
func generateBinPrinter(content []byte, contentType string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Add("Content-type", contentType)
		_, err := writer.Write(content)
		if err != nil {
			fmt.Printf("error writing constant binary string: %s\n", err)
		}
	}
}

//SetIcon creates a go handler in h that will return an icon to be displayed in response to /favicon.ico.
//The binaryIcon should be an array of bytes (usually created via 'seven5tool embedfile')
func SetIcon(mux *http.ServeMux, binaryIcon []byte) {
	mux.HandleFunc("/favicon.ico", generateBinPrinter(binaryIcon, "image/x-icon"))
}
