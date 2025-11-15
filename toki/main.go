package toki

import (
	"errors"
	"fmt"
	"time"

	"github.com/pratyushtiwary/toki/config"
	"github.com/pratyushtiwary/toki/log"
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

		Check(err, "StrategyCreationError")

		authStrategies = append(authStrategies, authStrategy)
	}

	for idx, authStrategy := range authStrategies {
		fmt.Printf("Running: `%s` for auth refresh\n", authComands[idx].Params.Command)
		defer authStrategy.Cleanup()
		err := authStrategy.Refresh(true)
		Check(err, "AuthRefreshError")
	}

	fmt.Printf("Done with initial auth refresh, executing `%s`\n", execCommand)

	// now that initial auth refresh is done, we'll start with main thread block
	command, err := process.NewCommand(execCommand, nil)

	Check(err, "MainCommandCreationError")

	err = command.Run(nil)

	Check(err, "MainCommandExecutionError")

	defer command.Cleanup()

	go func() {
		command.GetCmd().Wait()
	}() // this will make sure we don't create zombie process

	for {
		running, err := command.IsRunning()

		Check(err, "MainCommandStatusCheckError")

		if !running {
			break
		}

		for _, authStrategy := range authStrategies {

			expired, err := authStrategy.IsExpired()

			Check(err, "AuthExpiryCheckError")

			if expired {
				fmt.Print("Suspending main process\n")
				command.Suspend()
				fmt.Print("Refreshing auth\n")
				err := authStrategy.Refresh(false)

				Check(err, "AuthRefreshError")
				fmt.Print("Auth refreshed successfully, resuming main process\n")

				command.Resume()
			}
		}

		time.Sleep(time.Duration(1) * time.Second)
	}

	// check exit status of main process
	if !command.GetCmd().ProcessState.Success() {
		// stop and exit
		errorStr := command.GetStderrBuffer().String()
		Check(errors.New(errorStr), "MainCommandFailedError")
	}
	log.Info("Main process exited successfully")
}

func Run(cmd *cobra.Command, args []string) {
	if len(args) <= 0 {
		panic("Please provide a valid pipeline config path")
	}

	pipelineFile := args[0]

	log.Info(TokiStr)
	log.Info("Parsing file: %s\n", pipelineFile)

	steps, err := config.NewPipelineConfig(pipelineFile)

	Check(err, "ParsingError")

	log.Info("%d step(s) loaded, executing step(s) sequentially\n", len(steps))
	for idx, step := range steps {
		log.Log("\n----------------- Executing step %d -------------------\n", idx+1)
		executeStep(&step) // blocking method, would block the loop not until this step's command execution is finished
		log.Log("-------------------------------------------------------\n")
	}
}
