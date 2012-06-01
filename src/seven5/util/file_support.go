package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

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
	//we have the line in final but we may need to rest file pointer
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
