package toki

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pratyushtiwary/toki/config"
	"github.com/pratyushtiwary/toki/log"
	"github.com/pratyushtiwary/toki/process"
	"github.com/pratyushtiwary/toki/server"
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

func executeStep(step *config.PipelineStepConfig, server *server.Server) {
	execCommand := step.Command
	authCommands := step.AuthCommands

	fmt.Printf("Starting auth refresh before running `%s` command\n", execCommand)

	var authStrategies = make([]strategy.Strategy, 0)

	for idx := range authCommands {
		authCommand := authCommands[idx]
		authStrategy, authCommand, err := strategy.NewStrategy(authCommand.StrategyName, authCommand, nil, step.GetProjectStepConfig())

		Check(err, "StrategyCreationError")

		authStrategies = append(authStrategies, authStrategy)
		authCommands[idx] = authCommand
	}

	for idx, authStrategy := range authStrategies {
		fmt.Printf("Running: `%s` for auth refresh\n", authCommands[idx].Params.Command)
		defer authStrategy.Cleanup()
		err := authStrategy.Refresh(true)
		Check(err, "AuthRefreshError")
	}

	fmt.Printf("Done with initial auth refresh, executing `%s`\n", execCommand)

	// now that initial auth refresh is done, we'll start with main thread block
	command, err := process.NewCommand(execCommand, nil)
	Check(err, "MainCommandCreationError")

	passkey, err := uuid.NewUUID()
	Check(err, "PasskeyGenerationError")
	server.SetPasskey(passkey.String())

	err = command.Run(nil, []string{
		fmt.Sprintf("TOKI_PORT=%d", server.GetPort()),
		fmt.Sprintf("TOKI_TOKEN=%s", passkey.String()),
	})

	Check(err, "MainCommandExecutionError")

	defer command.Cleanup()

	go func() {
		command.GetCmd().Wait()
	}() // this will make sure we don't create zombie process

	anyExpired := false
	for {
		anyExpired = false
		running, err := command.IsRunning()

		Check(err, "MainCommandStatusCheckError")

		if !running {
			break
		}

		for _, authStrategy := range authStrategies {
			expired, err := authStrategy.IsExpired()
			Check(err, "AuthExpiryCheckError")

			if expired {
				anyExpired = true
				break
			}
		}

		if anyExpired {
			fmt.Print("Suspending main process\n")
			command.Suspend()
			for _, authStrategy := range authStrategies {
				fmt.Print("Refreshing auth\n")
				err := authStrategy.Refresh(false)
				Check(err, "AuthRefreshError")
			}
			fmt.Print("Auth refreshed successfully, resuming main process\n")
			go server.Broadcast("AUTH_REFRESHED")
			command.Resume()
		}

		if command.GetCmd().ProcessState != nil && command.GetCmd().ProcessState.Exited() {
			break
		}

		time.Sleep(time.Duration(1) * time.Second)
	}

	go server.Broadcast("FINISHED")

	if !command.GetCmd().ProcessState.Success() {
		errorStr := command.GetStderrBuffer().String()
		Check(errors.New(errorStr), "MainCommandFailedError")
	}
	log.Info("Main process exited successfully")
}

func Run(cmd *cobra.Command, args []string) {
	if len(args) <= 0 {
		panic("Please provide a valid pipeline config path")
	}

	projectConfigPath, err := cmd.Flags().GetString("project-config")

	Check(err, "InvalidProjectConfig")

	inDevMode, err := cmd.Flags().GetBool("dev")

	Check(err, "DevArgError")

	pipelineFile := args[0]

	log.Info(TokiStr)
	log.Info("Parsing file: %s\n", pipelineFile)

	projectConfig, err := config.NewProjectConfig(projectConfigPath)

	Check(err, "ProjectConfigParsingError")

	steps, err := config.NewPipelineConfig(pipelineFile, projectConfig)

	Check(err, "PipelineConfigParsingError")

	log.Info("%d step(s) loaded, executing step(s) sequentially\n", len(steps))

	server := server.NewServer(3110, inDevMode)

	go server.Listen()

	defer server.Close()

	for idx, step := range steps {
		log.Log("\n----------------- Executing step %d -------------------\n", idx+1)
		executeStep(step, server) // blocking method, would block the loop not until this step's command execution is finished
		log.Log("-------------------------------------------------------\n")
	}
}
