package cmd

import (
	"github.com/spf13/cobra"
)

var fetchAllCmd = &cobra.Command{
	Use:   "fetch-all",
	Short: "Fetch emails, profile pics, and attachments in one pass",
	RunE:  fetchCombined,
}

var (
	commonApiToken string
)

func init() {

	fetchEmailsCmd.PersistentFlags().StringVar(&commonApiToken, "token", "", "Slack API token. Can be obtained here: https://api.slack.com/docs/oauth-test-tokens")
	fetchEmailsCmd.MarkPersistentFlagRequired("apitoken")

	emailsApiToken = commonApiToken
	attachmentsApiToken = commonApiToken

}
