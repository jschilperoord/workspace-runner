package main

import (
	"context"
	"log"

	"github.com/caarlos0/env"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
)

type config struct {
	TFE_TOKEN string `env:"TFE_TOKEN"`
	TFE_ORG   string `env:"TFE_ORG" envDefault:"cbh"`
}

type cliBuilder struct {
	rootCmd                  *cobra.Command
	tfeClient                *tfe.Client
	tfeConfig                config
	workspaceRunCmdBaseline  *cobra.Command
	workspaceRunCmdInception *cobra.Command
	workspaceRunCmdCustom    *cobra.Command
}

func (c *cliBuilder) addRootCmd() *cliBuilder {
	c.rootCmd = &cobra.Command{
		Use:   "cmd",
		Short: "CLI execute runs on specific sets of terraform cloud workspaces",
	}
	c.rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	return c
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
	config := &tfe.Config{
		Token:             cfg.TFE_TOKEN,
		RetryServerErrors: true,
	}
	client, err := tfe.NewClient(config)

	if err != nil {
		log.Fatal(err)
		return
	}

	cliBuilder := newCliBuilder(client, cfg)
	cliBuilder.addRootCmd().addWorkspaceRunCmdBaseline().addWorkspaceRunCmdInception().addWorkspaceRunCmdCustom().execute()
}

func (c *cliBuilder) runWorkspaceCmd(wildcard string) {
	workspaces, err := c.tfeClient.Workspaces.List(context.Background(), c.tfeConfig.TFE_ORG, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Loop over the total amount of pages in workspaces.TotalPages
	for i := 1; i <= workspaces.TotalPages; i++ {
		// Create a new request with the current page
		workspaces, err := c.tfeClient.Workspaces.List(context.Background(), c.tfeConfig.TFE_ORG, &tfe.WorkspaceListOptions{
			ListOptions:  tfe.ListOptions{PageNumber: i},
			WildcardName: wildcard,
		})
		if err != nil {
			log.Fatal(err)
		}
		// Loop over the workspaces
		for _, workspace := range workspaces.Items {
			log.Printf("Workspace: %s", workspace.Name)
			_, err := c.tfeClient.Runs.Create(context.Background(), tfe.RunCreateOptions{
				Workspace: &tfe.Workspace{ID: workspace.ID},
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (c *cliBuilder) addWorkspaceRunCmdBaseline() *cliBuilder {
	c.workspaceRunCmdBaseline = &cobra.Command{
		Use:   "baseline",
		Short: "Execute a run on all baseline workspaces",

		Run: func(_ *cobra.Command, _ []string) {
			c.runWorkspaceCmd("baseline-*")
		},
	}
	c.rootCmd.AddCommand(c.workspaceRunCmdBaseline)

	return c
}

func (c *cliBuilder) addWorkspaceRunCmdCustom() *cliBuilder {
	var filter string
	c.workspaceRunCmdCustom = &cobra.Command{
		Use:   "custom",
		Short: "Execute a run on all custom workspaces",

		Run: func(_ *cobra.Command, _ []string) {
			c.runWorkspaceCmd(filter + "*")
		},
	}
	// define required local flag
	c.workspaceRunCmdCustom.Flags().StringVarP(&filter, "filter", "b", "", "Use this filter to select workspaces")
	c.workspaceRunCmdCustom.MarkFlagRequired("filter")
	c.rootCmd.AddCommand(c.workspaceRunCmdCustom)

	return c

}

func (c *cliBuilder) addWorkspaceRunCmdInception() *cliBuilder {
	c.workspaceRunCmdInception = &cobra.Command{
		Use:   "inception",
		Short: "Execute a run on all inception workspaces",

		Run: func(_ *cobra.Command, _ []string) {
			c.runWorkspaceCmd("wl-inception-*")
		},
	}
	c.rootCmd.AddCommand(c.workspaceRunCmdInception)

	return c
}

func (c *cliBuilder) execute() error {
	return c.rootCmd.Execute()
}

func newCliBuilder(tfeClient *tfe.Client, tfeConfig config) *cliBuilder {
	return &cliBuilder{
		tfeClient: tfeClient,
		tfeConfig: tfeConfig,
	}
}
