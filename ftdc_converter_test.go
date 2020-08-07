package ftdc

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//questions:
//reading directory in test
//appending operation names
//timestamp conversion and testing
//writing out to csv
// spikes in metrics


func TestTranslation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//opens directory containing genny ftdc files in repo
	cwd,_:=os.Getwd()
	dirPath:=cwd+"/Test-ftdc-files/"
	d, err := os.Open(dirPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, file := range files {

		if file.Mode().IsRegular() {

			completeFilePath := dirPath + file.Name()
			if (filepath.Ext(completeFilePath) == ".ftdc") {
				f, _ := os.Open(completeFilePath)
				//extract genny operation name from filename
				operationName:=strings.Split(file.Name(),".")[1]
				iter := ReadChunks(ctx, f)
				//out := &bytes.Buffer{}
				err := WriteGennyCSV(ctx, iter, os.Stdout,operationName )
				require.NoError(t, err)

			}
		}

	}


} //end of TestTranslation




