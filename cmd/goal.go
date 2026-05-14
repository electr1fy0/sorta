package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/electr1fy0/sorta/internal/agent"
	"github.com/electr1fy0/sorta/internal/llm"
	"github.com/spf13/cobra"
)

var (
	goalModel     string
	goalMaxSteps  int
	goalNames     []string
	goalNamesFile string
)

var goalCmd = &cobra.Command{
	Use:   "agent <directory> <instruction>",
	Short: "Plan and execute a natural language file-organization goal",
	Aliases: []string{
		"goal",
	},
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}

		names, err := loadAgentNames(goalNames, goalNamesFile)
		if err != nil {
			return err
		}

		client, err := llm.NewClient(goalModel)
		if err != nil {
			return err
		}

		session := agent.NewSession(client, agent.SessionOptions{MaxSteps: goalMaxSteps})
		for range goalMaxSteps {
			plan, err := session.Plan(context.Background(), dir, args[1], goalModel, names)
			if err != nil {
				return err
			}

			if len(plan.RequestFilenames) > 0 {
				fmt.Println(strings.TrimSpace(plan.Summary))
				extraNames, err := collectAgentNames(plan.RequestFilenames)
				if err != nil {
					return err
				}
				names = append(names, extraNames...)
				continue
			}

			fmt.Println(plan.String())
			if dryRun {
				fmt.Println("Dry run complete. No changes made.")
				return nil
			}

			fmt.Print("Do you want to execute this plan? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			ans, _ := reader.ReadString('\n')
			ans = strings.ToLower(strings.TrimSpace(ans))
			if ans != "y" && ans != "yes" {
				fmt.Println("Operation cancelled.")
				return nil
			}

			return agent.ExecutePlan(context.Background(), dir, plan)
		}

		return fmt.Errorf("planner hit max clarification rounds (%d)", goalMaxSteps)
	},
}

func init() {
	goalCmd.Flags().StringVar(&goalModel, "model", llm.DefaultModel, "Planner model to use")
	goalCmd.Flags().IntVar(&goalMaxSteps, "max-steps", 8, "Maximum number of planner tool-call rounds")
	goalCmd.Flags().StringSliceVar(&goalNames, "names", nil, "Extra labels or hints to send to the planner")
	goalCmd.Flags().StringVar(&goalNamesFile, "names-file", "", "File containing one extra planner hint per line")
	rootCmd.AddCommand(goalCmd)
}
