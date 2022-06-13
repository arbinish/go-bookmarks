package cmd

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new bookmark",
	Long: `
	Create a new bookmark entry. Expects a name, url, and tags for quick search. `,
	Run: func(cmd *cobra.Command, args []string) {
		name := cmd.Flag("name").Value.String()
		tags := cmd.Flag("tags").Value.String()
		url := cmd.Flag("url").Value.String()
		client := newClient("http://localhost:4912", 5)
		if client.create(name, url, tags) {
			fmt.Println("created")
			return
		}
		fmt.Printf("failed to create bookmark: %s\n", name)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	// Here you will define your flags and configuration settings.
	newCmd.PersistentFlags().String("url", "", "URL to save")
	newCmd.PersistentFlags().String("tags", "", "A comma separated list of tags for the given URL")
	newCmd.PersistentFlags().String("name", "", "A short name to refer the bookmark")
	newCmd.MarkPersistentFlagRequired("url")
	newCmd.MarkPersistentFlagRequired("tags")
	newCmd.MarkPersistentFlagRequired("name")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
