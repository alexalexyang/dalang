package util

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func IfFatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func GetCwd() (*string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &cwd, nil
}

func GetProjectRootDir() (*string, error) {

	// filename is the path of the file executing this function
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("failed to get current file path")
	}
	log.Println("Path of file executing this function: ", filename)

	projectDir := filepath.Join(filepath.Dir(filename), "..")

	return &projectDir, nil
}

func ExtractTarGz(gzipStream io.Reader, dirPath string) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatalf("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			fileTarget := filepath.Join(dirPath, header.Name)
			if err := os.Mkdir(fileTarget, 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			fileTarget := filepath.Join(dirPath, header.Name)
			err := os.MkdirAll(filepath.Dir(fileTarget), 0770)
			if err != nil {
				log.Fatalf("ExtractTarGz: MkDirAll() failed: %s", err.Error())
			}
			outFile, err := os.Create(fileTarget)
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
			}

			stats, err := os.Stat(fileTarget)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("File permissions: %s\n", stats.Mode())
			os.Chmod(fileTarget, fs.FileMode(header.Mode))

			stats, err = os.Stat(fileTarget)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("File permissions after Chmod: %s\n", stats.Mode())
		default:
			log.Fatalf(
				"ExtractTarGz: uknown type: %b in %s",
				header.Typeflag,
				header.Name)
		}
	}
}
