package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
)

func main() {
	if os.Getenv("INPUT_DEPS") == "" {
		fmt.Printf("INPUT_DEPS not set, nothing to do here ...")
		os.Exit(0)
	}

	depURLs := strings.Split(os.Getenv("INPUT_DEPS"), ",")

	tmpDir, err := ioutil.TempDir("", "dependency-action")
	logAndExitOnErr(fmt.Sprintf("unable to create tmp dir at '%s'", tmpDir), err)
	defer os.RemoveAll(tmpDir)

	os.Mkdir("bin")

	for _, depURL := range depURLs {
		resp, err := http.Get(depURL)
		logAndExitOnErr(fmt.Sprintf("unable to download file at '%s'", depURL), err)
		defer resp.Body.Close()

		tmpFile, err := ioutil.TempFile(tmpDir, "dependency")
		logAndExitOnErr(fmt.Sprintf("unable to create tmp file at '%s'", tmpFile.Name()), err)

		_, err = io.Copy(tmpFile, resp.Body)
		logAndExitOnErr(fmt.Sprintf("unable to copy to file at '%s'", tmpFile.Name()), err)

		unarchiver := configureUnarchiver(depURL)
		err = unarchiver.Unarchive(tmpFile.Name(), "/github/home/bin")
		logAndExitOnErr(fmt.Sprintf("unable to unarchive file at '%s'", tmpFile.Name()), err)
	}

	os.Remove("bin/ion-qt")
	os.Remove("bin/ion-tx")
	os.Remove("bin/test_ion")
}

func configureUnarchiver(depURL string) archiver.Unarchiver {
	switch extension := filepath.Ext(depURL); extension {
	case ".tgz":
		return archiver.DefaultTarGz
	case ".gz":
		return archiver.DefaultTarGz
	case ".txz":
		return archiver.DefaultTarXz
	case ".xz":
		return archiver.DefaultTarXz
	default:
		err := errors.New("unspported filetype")
		logAndExitOnErr(fmt.Sprintf("unable to unarchive file at '%s', ensure it is a supported filetype", depURL), err)
	}

	return nil
}

func logAndExitOnErr(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: "+msg)
		os.Exit(1)
	}
}
