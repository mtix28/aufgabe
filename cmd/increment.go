package cmd

import (
	"strconv"

	connection "example.de/demo/db"
	"github.com/spf13/cobra"
)

var incrementCmd = &cobra.Command{
	Use:   "increment [x]",
	Short: "increment",
	Long:  "increment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		value := args[0]
		incQuery := `UPDATE "Data" SET "IncValue" = "IncValue" + $1`
		i, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}

		connection.UpdateQuery(incQuery, i)
	},
}
