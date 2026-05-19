package cmd

import (
	"strconv"

	connection "example.de/demo/db"
	"github.com/spf13/cobra"
)

var decrementCmd = &cobra.Command{
	Use:   "decrement [x]",
	Short: "decrement",
	Long:  "decrement",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		value := args[0]
		decQuery := `UPDATE "Data" SET "DecValue" = "DecValue" - $1`
		i, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}

		connection.UpdateQuery(decQuery, i)
	},
}
