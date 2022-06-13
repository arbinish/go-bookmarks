/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete bookmark entry",
	Long: `Deletes an already added bookmark entry.
`,
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient("http://localhost:4912", 5)
		name := cmd.Flag("name").Value.String()
		if client.delete(name) {
			fmt.Println("deleted")
			return
		}
		fmt.Printf("%s: failed to delete\n", name)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")
	deleteCmd.PersistentFlags().String("name", "", "bookmark name to delete")

	deleteCmd.MarkPersistentFlagRequired("name")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
