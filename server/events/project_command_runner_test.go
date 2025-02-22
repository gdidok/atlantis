// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	tmocks "github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	eventmocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that it runs the expected plan steps.
func TestDefaultProjectCommandRunner_Plan(t *testing.T) {
	RegisterMockTestingT(t)
	mockInit := mocks.NewMockStepRunner()
	mockPlan := mocks.NewMockStepRunner()
	mockApply := mocks.NewMockStepRunner()
	mockRun := mocks.NewMockCustomStepRunner()
	realEnv := runtime.EnvStepRunner{}
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockApplyReqHandler := mocks.NewMockApplyRequirement()

	runner := events.DefaultProjectCommandRunner{
		Locker:                     mockLocker,
		LockURLGenerator:           mockURLGenerator{},
		InitStepRunner:             mockInit,
		PlanStepRunner:             mockPlan,
		ApplyStepRunner:            mockApply,
		RunStepRunner:              mockRun,
		EnvStepRunner:              &realEnv,
		PullApprovedChecker:        nil,
		WorkingDir:                 mockWorkingDir,
		Webhooks:                   nil,
		WorkingDirLocker:           events.NewDefaultWorkingDirLocker(),
		AggregateApplyRequirements: mockApplyReqHandler,
	}

	repoDir, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.Clone(
		matchers.AnyPtrToLoggingSimpleLogger(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyString(),
	)).ThenReturn(repoDir, false, nil)
	When(mockLocker.TryLock(
		matchers.AnyPtrToLoggingSimpleLogger(),
		matchers.AnyModelsPullRequest(),
		matchers.AnyModelsUser(),
		AnyString(),
		matchers.AnyModelsProject(),
	)).ThenReturn(&events.TryLockResponse{
		LockAcquired: true,
		LockKey:      "lock-key",
	}, nil)

	expEnvs := map[string]string{
		"name": "value",
	}
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		Steps: []valid.Step{
			{
				StepName:    "env",
				EnvVarName:  "name",
				EnvVarValue: "value",
			},
			{
				StepName: "run",
			},
			{
				StepName: "apply",
			},
			{
				StepName: "plan",
			},
			{
				StepName: "init",
			},
		},
		Workspace:  "default",
		RepoRelDir: ".",
	}
	// Each step will output its step name.
	When(mockInit.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("init", nil)
	When(mockPlan.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("plan", nil)
	When(mockApply.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("apply", nil)
	When(mockRun.Run(ctx, "", repoDir, expEnvs)).ThenReturn("run", nil)
	res := runner.Plan(ctx)

	Assert(t, res.PlanSuccess != nil, "exp plan success")
	Equals(t, "https://lock-key", res.PlanSuccess.LockURL)
	Equals(t, "run\napply\nplan\ninit", res.PlanSuccess.TerraformOutput)
	expSteps := []string{"run", "apply", "plan", "init", "env"}
	for _, step := range expSteps {
		switch step {
		case "init":
			mockInit.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
		case "plan":
			mockPlan.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
		case "apply":
			mockApply.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
		case "run":
			mockRun.VerifyWasCalledOnce().Run(ctx, "", repoDir, expEnvs)
		}
	}
}

func TestProjectOutputWrapper(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		Steps: []valid.Step{
			{
				StepName: "plan",
			},
		},
		Workspace:  "default",
		RepoRelDir: ".",
	}

	cases := []struct {
		Description string
		Failure     bool
		Error       bool
		Success     bool
		CommandName command.Name
	}{
		{
			Description: "plan success",
			Success:     true,
			CommandName: command.Plan,
		},
		{
			Description: "plan failure",
			Failure:     true,
			CommandName: command.Plan,
		},
		{
			Description: "plan error",
			Error:       true,
			CommandName: command.Plan,
		},
		{
			Description: "apply success",
			Success:     true,
			CommandName: command.Apply,
		},
		{
			Description: "apply failure",
			Failure:     true,
			CommandName: command.Apply,
		},
		{
			Description: "apply error",
			Error:       true,
			CommandName: command.Apply,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			var prjResult command.ProjectResult
			var expCommitStatus models.CommitStatus

			mockJobURLSetter := eventmocks.NewMockJobURLSetter()
			mockJobMessageSender := eventmocks.NewMockJobMessageSender()
			mockProjectCommandRunner := mocks.NewMockProjectCommandRunner()

			runner := &events.ProjectOutputWrapper{
				JobURLSetter:         mockJobURLSetter,
				JobMessageSender:     mockJobMessageSender,
				ProjectCommandRunner: mockProjectCommandRunner,
			}

			if c.Success {
				prjResult = command.ProjectResult{
					PlanSuccess:  &models.PlanSuccess{},
					ApplySuccess: "exists",
				}
				expCommitStatus = models.SuccessCommitStatus
			} else if c.Failure {
				prjResult = command.ProjectResult{
					Failure: "failure",
				}
				expCommitStatus = models.FailedCommitStatus
			} else if c.Error {
				prjResult = command.ProjectResult{
					Error: errors.New("error"),
				}
				expCommitStatus = models.FailedCommitStatus
			}

			When(mockProjectCommandRunner.Plan(matchers.AnyModelsProjectCommandContext())).ThenReturn(prjResult)
			When(mockProjectCommandRunner.Apply(matchers.AnyModelsProjectCommandContext())).ThenReturn(prjResult)

			switch c.CommandName {
			case command.Plan:
				runner.Plan(ctx)
			case command.Apply:
				runner.Apply(ctx)
			}

			mockJobURLSetter.VerifyWasCalled(Once()).SetJobURLWithStatus(ctx, c.CommandName, models.PendingCommitStatus)
			mockJobURLSetter.VerifyWasCalled(Once()).SetJobURLWithStatus(ctx, c.CommandName, expCommitStatus)

			switch c.CommandName {
			case command.Plan:
				mockProjectCommandRunner.VerifyWasCalledOnce().Plan(ctx)
			case command.Apply:
				mockProjectCommandRunner.VerifyWasCalledOnce().Apply(ctx)
			}
		})
	}
}

// Test what happens if there's no working dir. This signals that the project
// was never planned.
func TestDefaultProjectCommandRunner_ApplyNotCloned(t *testing.T) {
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir: mockWorkingDir,
	}
	ctx := command.ProjectContext{}
	When(mockWorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn("", os.ErrNotExist)

	res := runner.Apply(ctx)
	ErrEquals(t, "project has not been cloned–did you run plan?", res.Error)
}

// Test that if approval is required and the PR isn't approved we give an error.
func TestDefaultProjectCommandRunner_ApplyNotApproved(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:       mockWorkingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		AggregateApplyRequirements: &events.AggregateApplyRequirements{
			WorkingDir: mockWorkingDir,
		},
	}
	ctx := command.ProjectContext{
		ApplyRequirements: []string{"approved"},
	}
	tmp, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(tmp, nil)

	res := runner.Apply(ctx)
	Equals(t, "Pull request must be approved by at least one person other than the author before running apply.", res.Failure)
}

// Test that if mergeable is required and the PR isn't mergeable we give an error.
func TestDefaultProjectCommandRunner_ApplyNotMergeable(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:       mockWorkingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		AggregateApplyRequirements: &events.AggregateApplyRequirements{
			WorkingDir: mockWorkingDir,
		},
	}
	ctx := command.ProjectContext{
		PullReqStatus: models.PullReqStatus{
			Mergeable: false,
		},
		ApplyRequirements: []string{"mergeable"},
	}
	tmp, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(tmp, nil)

	res := runner.Apply(ctx)
	Equals(t, "Pull request must be mergeable before running apply.", res.Failure)
}

// Test that if undiverged is required and the PR is diverged we give an error.
func TestDefaultProjectCommandRunner_ApplyDiverged(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:       mockWorkingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		AggregateApplyRequirements: &events.AggregateApplyRequirements{
			WorkingDir: mockWorkingDir,
		},
	}
	ctx := command.ProjectContext{
		ApplyRequirements: []string{"undiverged"},
	}
	tmp, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(tmp, nil)

	res := runner.Apply(ctx)
	Equals(t, "Default branch must be rebased onto pull request before running apply.", res.Failure)
}

// Test that it runs the expected apply steps.
func TestDefaultProjectCommandRunner_Apply(t *testing.T) {
	cases := []struct {
		description string
		steps       []valid.Step
		applyReqs   []string

		expSteps      []string
		expOut        string
		expFailure    string
		pullMergeable bool
	}{
		{
			description: "normal workflow",
			steps:       valid.DefaultApplyStage.Steps,
			expSteps:    []string{"apply"},
			expOut:      "apply",
		},
		{
			description: "approval required",
			steps:       valid.DefaultApplyStage.Steps,
			applyReqs:   []string{"approved"},
			expSteps:    []string{"approve", "apply"},
			expOut:      "apply",
		},
		{
			description:   "mergeable required",
			steps:         valid.DefaultApplyStage.Steps,
			pullMergeable: true,
			applyReqs:     []string{"mergeable"},
			expSteps:      []string{"apply"},
			expOut:        "apply",
		},
		{
			description:   "mergeable required, pull not mergeable",
			steps:         valid.DefaultApplyStage.Steps,
			pullMergeable: false,
			applyReqs:     []string{"mergeable"},
			expSteps:      []string{""},
			expOut:        "",
			expFailure:    "Pull request must be mergeable before running apply.",
		},
		{
			description:   "mergeable and approved required",
			steps:         valid.DefaultApplyStage.Steps,
			pullMergeable: true,
			applyReqs:     []string{"mergeable", "approved"},
			expSteps:      []string{"approved", "apply"},
			expOut:        "apply",
		},
		{
			description: "workflow with custom apply stage",
			steps: []valid.Step{
				{
					StepName:    "env",
					EnvVarName:  "key",
					EnvVarValue: "value",
				},
				{
					StepName: "run",
				},
				{
					StepName: "apply",
				},
				{
					StepName: "plan",
				},
				{
					StepName: "init",
				},
			},
			expSteps: []string{"env", "run", "apply", "plan", "init"},
			expOut:   "run\napply\nplan\ninit",
		},
	}

	for _, c := range cases {
		if c.description != "workflow with custom apply stage" {
			continue
		}
		t.Run(c.description, func(t *testing.T) {
			RegisterMockTestingT(t)
			mockInit := mocks.NewMockStepRunner()
			mockPlan := mocks.NewMockStepRunner()
			mockApply := mocks.NewMockStepRunner()
			mockRun := mocks.NewMockCustomStepRunner()
			mockEnv := mocks.NewMockEnvStepRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()
			mockSender := mocks.NewMockWebhooksSender()
			applyReqHandler := &events.AggregateApplyRequirements{
				WorkingDir: mockWorkingDir,
			}

			runner := events.DefaultProjectCommandRunner{
				Locker:                     mockLocker,
				LockURLGenerator:           mockURLGenerator{},
				InitStepRunner:             mockInit,
				PlanStepRunner:             mockPlan,
				ApplyStepRunner:            mockApply,
				RunStepRunner:              mockRun,
				EnvStepRunner:              mockEnv,
				WorkingDir:                 mockWorkingDir,
				Webhooks:                   mockSender,
				WorkingDirLocker:           events.NewDefaultWorkingDirLocker(),
				AggregateApplyRequirements: applyReqHandler,
			}
			repoDir, cleanup := TempDir(t)
			defer cleanup()
			When(mockWorkingDir.GetWorkingDir(
				matchers.AnyModelsRepo(),
				matchers.AnyModelsPullRequest(),
				AnyString(),
			)).ThenReturn(repoDir, nil)

			ctx := command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				Steps:             c.steps,
				Workspace:         "default",
				ApplyRequirements: c.applyReqs,
				RepoRelDir:        ".",
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{
						IsApproved: true,
					},
					Mergeable: true,
				},
			}
			expEnvs := map[string]string{
				"key": "value",
			}
			When(mockInit.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("init", nil)
			When(mockPlan.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("plan", nil)
			When(mockApply.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("apply", nil)
			When(mockRun.Run(ctx, "", repoDir, expEnvs)).ThenReturn("run", nil)
			When(mockEnv.Run(ctx, "", "value", repoDir, make(map[string]string))).ThenReturn("value", nil)

			res := runner.Apply(ctx)
			Equals(t, c.expOut, res.ApplySuccess)
			Equals(t, c.expFailure, res.Failure)

			for _, step := range c.expSteps {
				switch step {
				case "init":
					mockInit.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
				case "plan":
					mockPlan.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
				case "apply":
					mockApply.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
				case "run":
					mockRun.VerifyWasCalledOnce().Run(ctx, "", repoDir, expEnvs)
				case "env":
					mockEnv.VerifyWasCalledOnce().Run(ctx, "", "value", repoDir, expEnvs)
				}
			}
		})
	}
}

// Test that it runs the expected apply steps.
func TestDefaultProjectCommandRunner_ApplyRunStepFailure(t *testing.T) {
	RegisterMockTestingT(t)
	mockApply := mocks.NewMockStepRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockSender := mocks.NewMockWebhooksSender()
	applyReqHandler := &events.AggregateApplyRequirements{
		WorkingDir: mockWorkingDir,
	}

	runner := events.DefaultProjectCommandRunner{
		Locker:                     mockLocker,
		LockURLGenerator:           mockURLGenerator{},
		ApplyStepRunner:            mockApply,
		WorkingDir:                 mockWorkingDir,
		WorkingDirLocker:           events.NewDefaultWorkingDirLocker(),
		AggregateApplyRequirements: applyReqHandler,
		Webhooks:                   mockSender,
	}
	repoDir, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.GetWorkingDir(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyString(),
	)).ThenReturn(repoDir, nil)

	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		Steps: []valid.Step{
			{
				StepName: "apply",
			},
		},
		Workspace:         "default",
		ApplyRequirements: []string{},
		RepoRelDir:        ".",
	}
	expEnvs := map[string]string{}
	When(mockApply.Run(ctx, nil, repoDir, expEnvs)).ThenReturn("apply", fmt.Errorf("something went wrong"))

	res := runner.Apply(ctx)
	Assert(t, res.ApplySuccess == "", "exp apply failure")

	mockApply.VerifyWasCalledOnce().Run(ctx, nil, repoDir, expEnvs)
}

// Test run and env steps. We don't use mocks for this test since we're
// not running any Terraform.
func TestDefaultProjectCommandRunner_RunEnvSteps(t *testing.T) {
	RegisterMockTestingT(t)
	tfClient := tmocks.NewMockClient()
	tfVersion, err := version.NewVersion("0.12.0")
	Ok(t, err)
	run := runtime.RunStepRunner{
		TerraformExecutor: tfClient,
		DefaultTFVersion:  tfVersion,
	}
	env := runtime.EnvStepRunner{
		RunStepRunner: &run,
	}
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()

	runner := events.DefaultProjectCommandRunner{
		Locker:           mockLocker,
		LockURLGenerator: mockURLGenerator{},
		RunStepRunner:    &run,
		EnvStepRunner:    &env,
		WorkingDir:       mockWorkingDir,
		Webhooks:         nil,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
	}

	repoDir, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.Clone(
		matchers.AnyPtrToLoggingSimpleLogger(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyString(),
	)).ThenReturn(repoDir, false, nil)
	When(mockLocker.TryLock(
		matchers.AnyPtrToLoggingSimpleLogger(),
		matchers.AnyModelsPullRequest(),
		matchers.AnyModelsUser(),
		AnyString(),
		matchers.AnyModelsProject(),
	)).ThenReturn(&events.TryLockResponse{
		LockAcquired: true,
		LockKey:      "lock-key",
	}, nil)

	ctx := command.ProjectContext{
		Log: logging.NewNoopLogger(t),
		Steps: []valid.Step{
			{
				StepName:   "run",
				RunCommand: "echo var=$var",
			},
			{
				StepName:    "env",
				EnvVarName:  "var",
				EnvVarValue: "value",
			},
			{
				StepName:   "run",
				RunCommand: "echo var=$var",
			},
			{
				StepName:   "env",
				EnvVarName: "dynamic_var",
				RunCommand: "echo dynamic_value",
			},
			{
				StepName:   "run",
				RunCommand: "echo dynamic_var=$dynamic_var",
			},
			// Test overriding the variable
			{
				StepName:    "env",
				EnvVarName:  "dynamic_var",
				EnvVarValue: "overridden",
			},
			{
				StepName:   "run",
				RunCommand: "echo dynamic_var=$dynamic_var",
			},
		},
		Workspace:  "default",
		RepoRelDir: ".",
	}
	res := runner.Plan(ctx)
	Assert(t, res.PlanSuccess != nil, "exp plan success")
	Equals(t, "https://lock-key", res.PlanSuccess.LockURL)
	Equals(t, "var=\n\nvar=value\n\ndynamic_var=dynamic_value\n\ndynamic_var=overridden\n", res.PlanSuccess.TerraformOutput)
}

type mockURLGenerator struct{}

func (m mockURLGenerator) GenerateLockURL(lockID string) string {
	return "https://" + lockID
}
