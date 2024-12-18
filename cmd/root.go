package cmd

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/cobra"
)

var (
	inputArchive  string
	outputArchive string
	verbose       bool
	httpClient    *http.Client
)

var rootCmd = &cobra.Command{
	Use:   "slack-advanced-exporter",
	Short: "The Slack Advanced Exporter ",
	Long: `The Slack Advanced Exporter is a tool for supplementing official data exports from Slack with the other bits
and pieces that these don't include.

Version: 0.5.1`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&inputArchive, "input-archive", "i", "", "the path to the Slack export archive which you wish to augment")
	rootCmd.PersistentFlags().StringVarP(&outputArchive, "output-archive", "o", "", "the path to which you would like the output archive to be written")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "print detailed information about what is happening while the command is executing")
	rootCmd.MarkPersistentFlagRequired("input-archive")
	rootCmd.MarkPersistentFlagRequired("output-archive")
	rootCmd.AddCommand(fetchAttachmentsCmd)
	rootCmd.AddCommand(fetchEmailsCmd)
	rootCmd.AddCommand(fetchPrivateChannelsCmd)
	rootCmd.AddCommand(fetchProfilePicturesCmd)
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10
	httpClient = retryClient.StandardClient() // *http.Client

}

func Execute() error {
	return rootCmd.Execute()
}
