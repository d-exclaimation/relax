package rpc

import (
	"fmt"
	"log"

	"d-exclaimation.me/relax/lib/async"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// WorkflowInOut is the input and output of a workflow step
type WorkflowInOut struct {
	In  *slack.WorkflowStepInputs
	Out *[]slack.WorkflowStepOutput
}

// WorkflowExecutionResult is the result of a workflow step execution
type (
	WorkflowExecutionResult interface {
		kind() string
	}
	WorkflowSuccessResult struct {
		Outputs map[string]string
	}
	WorkflowFailureResult struct {
		Message string
	}
)

func (WorkflowSuccessResult) kind() string { return "success" }
func (WorkflowFailureResult) kind() string { return "failure" }

// Workflow is a workflow step
type Workflow[C any] struct {
	callbackID string
	edit       func(e slack.InteractionCallback, ctx C) []slack.Block
	save       func(e slack.InteractionCallback, ctx C) WorkflowInOut
	execute    func(e *slackevents.WorkflowStepExecuteEvent, ctx C) WorkflowExecutionResult
}

// Step creates a new workflow step
func Step[C any](callbackID string) Workflow[C] {
	return Workflow[C]{
		callbackID: callbackID,
		edit: func(e slack.InteractionCallback, ctx C) []slack.Block {
			return []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject(
						slack.MarkdownType,
						fmt.Sprintf("Add %s workflow step", callbackID),
						false,
						false,
					),
					nil,
					nil,
				),
			}
		},
		save: func(e slack.InteractionCallback, ctx C) WorkflowInOut {
			return WorkflowInOut{}
		},
	}
}

// OnEdit set the edit callback for the workflow
func (s Workflow[C]) OnEdit(edit func(e slack.InteractionCallback, ctx C) []slack.Block) Workflow[C] {
	s.edit = edit
	return s
}

// OnSave set the save callback for the workflow
func (s Workflow[C]) OnSave(save func(e slack.InteractionCallback, ctx C) WorkflowInOut) Workflow[C] {
	s.save = save
	return s
}

// OnExecute set the execute callback for the workflow
func (s Workflow[C]) OnExecute(execute func(e *slackevents.WorkflowStepExecuteEvent, ctx C) WorkflowExecutionResult) Workflow[C] {
	s.execute = execute
	return s
}

// WorkflowsRouter is a router for workflows
type WorkflowsRouter[C any] struct {
	client *slack.Client
	steps  []Workflow[C]
}

// Workflows creates a new workflows router
func Workflows[C any](
	client *slack.Client,
	steps ...Workflow[C],
) WorkflowsRouter[C] {
	return WorkflowsRouter[C]{
		client: client,
		steps:  steps,
	}
}

// HandleAsync handles the workflow step execute event
func (r *WorkflowsRouter[C]) HandleAsync(
	event *slackevents.WorkflowStepExecuteEvent,
	ctx func() C,
) async.Task[async.Unit] {
	return async.New(func() (async.Unit, error) {
		for _, step := range r.steps {
			if step.callbackID == event.CallbackID {
				err := error(nil)
				switch result := step.execute(event, ctx()).(type) {
				case WorkflowSuccessResult:
					err = r.client.WorkflowStepCompleted(
						event.WorkflowStep.WorkflowStepExecuteID,
						slack.WorkflowStepCompletedRequestOptionOutput(
							result.Outputs,
						),
					)
				case WorkflowFailureResult:
					err = r.client.WorkflowStepFailed(
						event.WorkflowStep.WorkflowStepExecuteID,
						result.Message,
					)
				}

				if err != nil {
					log.Fatalln(err)
				}
				return async.Done, nil
			}
		}
		return async.Done, nil
	})
}

// HandleInteractionAsync handles the workflow step edit and save events
func (r *WorkflowsRouter[C]) HandleInteractionAsync(
	event slack.InteractionCallback,
	ctx func() C,
) async.Task[async.Unit] {
	return async.New(func() (async.Unit, error) {
		for _, step := range r.steps {
			if step.callbackID == event.CallbackID {
				err := error(nil)
				c := ctx()

				switch event.Type {
				case slack.InteractionTypeWorkflowStepEdit:
					res := step.edit(event, c)
					_, err = r.client.OpenView(
						event.TriggerID,
						slack.NewConfigurationModalRequest(
							slack.Blocks{
								BlockSet: res,
							},
							"",
							"",
						).ModalViewRequest,
					)

				case slack.InteractionTypeViewSubmission:
					res := step.save(event, c)
					err = r.client.SaveWorkflowStepConfiguration(
						event.WorkflowStep.WorkflowStepEditID,
						res.In,
						res.Out,
					)
				}

				if err != nil {
					log.Fatalln(err)
				}
				return async.Done, nil
			}
		}
		return async.Done, nil
	})
}
