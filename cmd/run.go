package cmd

import (
	"fmt"
	"os"

	"github.com/ArttuOll/go-csv/internal/parser"
	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string) error {

	filename := args[0]

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %v", filename)
	}

	defer file.Close()

	csvParser := parser.NewCSVParser(file)

	for {
		record, err := csvParser.Parse()
		if err != nil {
			fmt.Println(err)
			break
		}

		if record == nil {
			break
		}

		fmt.Println(record)
	}

	return nil
}
