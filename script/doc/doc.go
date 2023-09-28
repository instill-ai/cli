package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/instill-ai/cli/internal/build"
	"github.com/instill-ai/cli/pkg/cmd/factory"
	"github.com/instill-ai/cli/pkg/cmd/root"
)

var dir = "doc"

func main() {
	buildDate := build.Date
	buildVersion := build.Version
	cmdFactory := factory.New(buildVersion)
	runGenManCmd(root.NewCmdRoot(cmdFactory, buildVersion, buildDate))
}

func runGenManCmd(cmd *cobra.Command) error {
	header := &doc.GenManHeader{
		Section: "1",
		Manual:  "Instill AI Manual",
		// TOOD version
		Source: "Instill AI",
	}

	if !strings.HasSuffix(dir, string(os.PathSeparator)) {
		dir += string(os.PathSeparator)
	}

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err := doc.GenManTree(cmd.Root(), header, dir); err != nil {
		return err
	}

	fmt.Println("Generated Instill AI man pages in", dir)

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	_ = os.Chdir(dir)

	for _, file := range files {
		md, err := os.Create(file.Name() + ".md")
		if err != nil {
			log.Fatal(err)
		}

		cmd := exec.Command("man2md", file.Name())
		cmd.Stdout = md
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		md.Close()
	}

	fmt.Println("Generated Instill AI markdown pages in", dir)

	return nil
}
