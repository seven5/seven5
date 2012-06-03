package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//ReadIntoString reads a whole file into a string.  This is probably in the
//standard library.
func ReadIntoString(dir string, filename string) (string, error) {
	var file *os.File
	var info os.FileInfo
	var err error
	var buffer bytes.Buffer

	fpath := filepath.Join(dir, filename)
	if info, err = os.Stat(fpath); err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", errors.New(fmt.Sprintf("in %s, %s is a directory not a file!",
			dir, filename))
	}

	if file, err = os.Open(fpath); err != nil {
		return "", err
	}

	if _, err = buffer.ReadFrom(file); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

//ReadLine reads a single line, at current position, from a file and leaves
//the file pointed at the next line unless io.EOF is reached in which case
//the file pointer position is undefined.  It returns the string read and
//the error, both of which may be non-zero-valued if the error is io.EOF.
//Isn't this in the standard library?
func ReadLine(file *os.File) (string, error) {
	buff := make([]byte, 256)
	final :=[]byte{}
	
	x, _ := os.Create("/tmp/out")
	defer func() {
		x.Close()
	}()
	pos, err := file.Seek(0, 1)
	fmt.Fprintf(x,"seek position %d\n",pos)
	
	if err != nil {
		return "", err
	}
	done := false

	for {
		n, err := file.Read(buff)
		fmt.Fprintf(x,"bytes read %d\n",n)
		
		//EOF? put rest on this line 
		if err == io.EOF {
			final = append(final, buff[0:n]...)
			done = true
			break
		}
		//IO error?
		if err != nil {
			return "", err
		}

		//targ is where the CR is if found
		targ := -1
		for i := 0; i < n; i++ {
			fmt.Fprintf(x,"%v==10 %v  [%d]\n",buff[i],buff[i]==10,i)
		
			if buff[i] == '\n' {
				targ = i
				break
			}
		}
		fmt.Fprintf(x,"targ is %d and %s, %s\n",targ, string(buff), string(final))
		
		//no CR found, read some more?
		if targ == -1 {
			final = append(final, buff...)
			continue
		}
		//found it
		final = append(final, buff[0:targ]...)
		break
	}
	//we have the line in final but we may need to reset file pointer
	if !done {
		fmt.Fprintf(x,"seeking %d\n",pos+int64(len(final)))
	
		_, err = file.Seek(pos+int64(len(final))+1, 0)
		if err!=nil {
			return "",nil
		}
	} else {
		return string(final), io.EOF
	}
	return string(final), nil
}

//FilenameToTypeName is our algorithm for taking suffixed filenames and turning
//them into mixed case, exported type names.
func FilenameToTypeName(filename string) string {
	if !strings.HasSuffix(filename,".go") {
		panic("called FilenameToTypeName with a non-go filename")
	}
	f := strings.Replace(filename, ".go", "", 1)
	piece := strings.Split(f, "_")
	if len(piece) == 1 {
		panic("called FilenameToTypeName with a non underscored filename")
	}
	if len(piece) == 2 {
		return CapFirstLetter(piece[0])
	}
	result := []string{}
	
	for _, p:= range piece {
		result = append(result, CapFirstLetter(p))
	}
	return strings.Join(result, "")
	
}

//TypenameToFilename reverses the action of FilenameToTypename but doesn't
//deal with suffixes, just embedded underscores.
func TypeNameToFilename(filename string) string {
	var buffer bytes.Buffer
	
	inLowerSequence := true;
	let:=strings.Split(filename, "")
	for i, letter:= range let {
		if inLowerSequence && (strings.ToUpper(letter)==letter) {
			if i!=0 {
				buffer.WriteString("_")
			}
			buffer.WriteString(strings.ToLower(letter))
			inLowerSequence=false
			continue
		} 
		inLowerSequence = true
		buffer.WriteString(letter)
	}
	return buffer.String()
}

//CapFirstLetter makes sure the first letter of a string is capitalized.
func CapFirstLetter(s string) string {
	letter:=strings.Split(s,"")
	letter[0]= strings.ToUpper(letter[0])
	return strings.Join(letter,"")
}
