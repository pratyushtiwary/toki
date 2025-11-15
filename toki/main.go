package toki

import (
	"fmt"
	"time"

	"github.com/pratyushtiwary/toki/config"
	"github.com/pratyushtiwary/toki/process"
	"github.com/pratyushtiwary/toki/strategy"
	"github.com/spf13/cobra"
)

const TokiStr = `
 ______   ______     __  __     __    
/\__  _\ /\  __ \   /\ \/ /    /\ \   
\/_/\ \/ \ \ \/\ \  \ \  _"-.  \ \ \  
   \ \_\  \ \_____\  \ \_\ \_\  \ \_\ 
    \/_/   \/_____/   \/_/\/_/   \/_/ 

`

func executeStep(step *config.PipelineStepConfig) {
	execCommand := step.Command
	authComands := step.AuthCommands

	fmt.Printf("Starting auth refresh before running `%s` command\n", execCommand)

	var authStrategies = make([]strategy.Strategy, 0)

	for _, authCommand := range authComands {
		authStrategy, err := strategy.NewStrategy(authCommand.StrategyName, authCommand.Params, nil)

		Check(err)

		authStrategies = append(authStrategies, authStrategy)
	}

	for idx, authStrategy := range authStrategies {
		fmt.Printf("Running: `%s` for auth refresh\n", authComands[idx].Params.Command)
		defer authStrategy.Cleanup()
		authStrategy.Refresh(true)
	}

	fmt.Printf("Done with initial auth refresh, executing `%s`\n", execCommand)

	// now that initial auth refresh is done, we'll start with main thread block
	command, err := process.NewCommand(execCommand, nil)

	Check(err)

	err = command.Run(nil)

	Check(err)

	defer command.Cleanup()

	go func() {
		command.GetCmd().Wait()
	}() // this will make sure we don't create zombie process

	for {
		running, err := command.IsRunning()

		Check(err)

		if !running {
			break
		}

		for _, authStrategy := range authStrategies {

			expired, err := authStrategy.IsExpired()

			Check(err)

			if expired {
				fmt.Print("Suspending main process\n")
				command.Suspend()
				fmt.Print("Refreshing auth\n")
				err := authStrategy.Refresh(false)

				Check(err)
				fmt.Print("Auth refreshed successfully, resuming main process\n")

				command.Resume()
			}
		}

		time.Sleep(1)
	}
}

func Run(cmd *cobra.Command, args []string) {
	if len(args) <= 0 {
		panic("Please provide a valid pipeline config path")
	}

	pipelineFile := args[0]

	fmt.Print(TokiStr)
	fmt.Printf("Parsing file: %s\n", pipelineFile)

	steps, err := config.NewPipelineConfig(pipelineFile)

	Check(err)

	fmt.Printf("%d step(s) loaded, executing step(s) sequentially\n", len(steps))
	for idx, step := range steps {
		fmt.Printf("\n----------------- Executing step %d -------------------\n", idx+1)
		executeStep(&step) // blocking method, would block the loop not until this step's command execution is finished
		fmt.Printf("-------------------------------------------------------\n")
	}
}
