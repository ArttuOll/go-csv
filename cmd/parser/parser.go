package parser

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Parse(cmd *cobra.Command, args []string) (records [][]string, err error) {
	filename := args[0]

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %v", filename)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return nil, nil
}
