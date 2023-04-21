package pulumiInstaller

import (
	"dalang/config"
	"dalang/util"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// Checks if Pulumi is already installed.
// If no, downloads Pulumi tar.gz, extracts it, and deletes the tar.gz.
// Also sets path to extracted Pulumi binary.
func GetPulumi() error {
	projDir := config.Config.ProjectRootDir

	dirPath := path.Join(projDir, ".pulumi")
	pulumiPath := path.Join(dirPath, "pulumi")

	_, err := os.Stat(pulumiPath)

	if !errors.Is(err, os.ErrNotExist) {
		log.Println("Found: Pulumi binary; setting path, skipping download")

		err := setPulumiPath(pulumiPath)
		util.IfFatalErr(err)

		return nil
	}

	log.Println("Not found: Pulumi binary. Installing Pulumi")

	pulumiRelease := config.Config.PulumiReleaseUrl

	response, err := http.Get(pulumiRelease)

	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return errors.New("failed to download Pulumi")
	}

	defer response.Body.Close()

	filePath := path.Join(projDir, "pulumi.tar.gz")

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer out.Close()
	_, err = io.Copy(out, response.Body)
	if err != nil {
		return err
	}

	openedFile, err := os.Open(out.Name())
	if err != nil {
		return err
	}

	util.ExtractTarGz(openedFile, dirPath)

	setPulumiPath(pulumiPath)

	err = os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}

func setPulumiPath(pulumiPath string) error {
	return os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), pulumiPath))
}

func GetPulumiWorkspaceBackendDir() (*string, error) {
	projDir := config.Config.ProjectRootDir

	workspaceBackendPath := filepath.Join(projDir, "pulumi-backend")

	_, err := os.Stat(workspaceBackendPath)

	if errors.Is(err, os.ErrNotExist) {
		log.Println("Not found: Pulumi workspace backend directory. Creating: ", workspaceBackendPath)

		err := os.Mkdir(workspaceBackendPath, os.ModePerm)
		util.IfFatalErr(err)

		return nil, err
	}

	log.Println("Found: Pulumi workspace backend directory: ", workspaceBackendPath)

	return &workspaceBackendPath, nil
}
