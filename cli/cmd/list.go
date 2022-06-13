package cmd

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/

import (
	"fmt"

	"github.com/arbinish/go-bookmarks/bookmarks"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List bookmark by name",
	Long: `
	List bookmarks by name or tags.
`,
	Run: func(cmd *cobra.Command, args []string) {
		name := cmd.Flag("name").Value.String()
		tag := cmd.Flag("tag").Value.String()
		client := newClient("http://localhost:4912", 5)
		if name == "" && tag == "" {
			fmt.Println(cmd.UsageString())
			return
		}
		var r []*bookmarks.Bookmark
		var param string
		if name != "" {
			r = client.findByParam("name", name)
			param = name
		}
		if tag != "" {
			r = client.findByParam("tag", tag)
			param = tag
		}
		if r == nil {
			fmt.Printf("%s: not found\n", param)
			return
		}
		for _, p := range r {
			fmt.Printf("%s\n", p)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	listCmd.PersistentFlags().String("name", "", "short name to look up")
	listCmd.PersistentFlags().String("tag", "", "tag to search for")
}
