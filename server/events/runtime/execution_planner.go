package runtime

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/logging"
)

const PlanStageName = "plan"
const ApplyStageName = "apply"
const AtlantisYAMLFilename = "atlantis.yaml"

type ExecutionPlanner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
	ParserValidator   *yaml.ParserValidator
}

type TerraformExec interface {
	RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error)
}

func (s *ExecutionPlanner) BuildPlanStage(log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string) (*PlanStage, error) {
	defaults := s.defaultPlanSteps(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
	steps, err := s.buildStage(PlanStageName, log, repoDir, workspace, relProjectPath, extraCommentArgs, username, defaults)
	if err != nil {
		return nil, err
	}
	return &PlanStage{
		Steps: steps,
	}, nil
}

func (s *ExecutionPlanner) buildStage(stageName string, log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string, defaults []Step) ([]Step, error) {
	config, err := s.ParserValidator.ReadConfig(repoDir)

	// If there's no config file, use defaults.
	if os.IsNotExist(err) {
		log.Info("no %s file found––continuing with defaults", AtlantisYAMLFilename)
		return defaults, nil
	}

	if err != nil {
		return nil, err
	}

	// Get this project's configuration.
	for _, p := range config.Projects {
		if p.Dir == relProjectPath && p.Workspace == workspace {
			workflowName := p.Workflow

			// If they didn't specify a workflow, use the default.
			if workflowName == "" {
				log.Info("no %s workflow set––continuing with defaults", AtlantisYAMLFilename)
				return defaults, nil
			}

			// If they did specify a workflow, find it.
			workflow, exists := config.Workflows[workflowName]
			if !exists {
				return nil, fmt.Errorf("no workflow with key %q defined", workflowName)
			}

			// We have a workflow defined, so now we need to build it.
			meta := s.buildMeta(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
			var steps []Step
			var stepsConfig []yaml.StepConfig
			if stageName == PlanStageName {
				stepsConfig = workflow.Plan.Steps
			} else {
				stepsConfig = workflow.Apply.Steps
			}
			for _, stepConfig := range stepsConfig {
				var step Step
				switch stepConfig.StepType {
				case "init":
					step = &InitStep{
						Meta:      meta,
						ExtraArgs: stepConfig.ExtraArgs,
					}
				case "plan":
					step = &PlanStep{
						Meta:      meta,
						ExtraArgs: stepConfig.ExtraArgs,
					}
				case "apply":
					step = &ApplyStep{
						Meta:      meta,
						ExtraArgs: stepConfig.ExtraArgs,
					}
				case "run":
					step = &RunStep{
						Meta:     meta,
						Commands: stepConfig.Run,
					}
				}
				steps = append(steps, step)
			}
			return steps, nil
		}
	}
	// They haven't defined this project, use the default workflow.
	log.Info("no project with dir %q and workspace %q defined; continuing with defaults", relProjectPath, workspace)
	return defaults, nil
}

func (s *ExecutionPlanner) BuildApplyStage(log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string) (*ApplyStage, error) {
	defaults := s.defaultApplySteps(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
	steps, err := s.buildStage(ApplyStageName, log, repoDir, workspace, relProjectPath, extraCommentArgs, username, defaults)
	if err != nil {
		return nil, err
	}
	return &ApplyStage{
		Steps: steps,
	}, nil
}

func (s *ExecutionPlanner) buildMeta(log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string) StepMeta {
	return StepMeta{
		Log:                   log,
		Workspace:             workspace,
		AbsolutePath:          filepath.Join(repoDir, relProjectPath),
		DirRelativeToRepoRoot: relProjectPath,
		// If there's no config then we should use the default tf version.
		TerraformVersion:  s.DefaultTFVersion,
		TerraformExecutor: s.TerraformExecutor,
		ExtraCommentArgs:  extraCommentArgs,
		Username:          username,
	}
}

func (s *ExecutionPlanner) defaultPlanSteps(log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string) []Step {
	meta := s.buildMeta(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
	return []Step{
		&InitStep{
			ExtraArgs: nil,
			Meta:      meta,
		},
		&PlanStep{
			ExtraArgs: nil,
			Meta:      meta,
		},
	}
}
func (s *ExecutionPlanner) defaultApplySteps(log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string) []Step {
	meta := s.buildMeta(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
	return []Step{
		&ApplyStep{
			ExtraArgs: nil,
			Meta:      meta,
		},
	}
}