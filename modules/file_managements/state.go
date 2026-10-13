package file_managements

import (
	"context"
	"fmt"
	"sort"

	"github.com/looplab/fsm"

	"go-api/constant"
)

// FSM events of a file management job, see docs/prd/file-managements.md §5.3
const (
	EventProcess  = "process"
	EventComplete = "complete"
	EventFail     = "fail"
	EventCancel   = "cancel"
	EventRetry    = "retry"
)

const (
	stateNew        = "new"
	stateOnProgress = "on_progress"
	stateFailed     = "failed"
	stateCancelled  = "cancelled"
	stateCompleted  = "completed"
)

var statusStates = map[int8]string{
	constant.FileManagementNew:        stateNew,
	constant.FileManagementOnProgress: stateOnProgress,
	constant.FileManagementFailed:     stateFailed,
	constant.FileManagementCancelled:  stateCancelled,
	constant.FileManagementCompleted:  stateCompleted,
}

var stateStatuses = map[string]int8{
	stateNew:        constant.FileManagementNew,
	stateOnProgress: constant.FileManagementOnProgress,
	stateFailed:     constant.FileManagementFailed,
	stateCancelled:  constant.FileManagementCancelled,
	stateCompleted:  constant.FileManagementCompleted,
}

var stateEvents = fsm.Events{
	{Name: EventProcess, Src: []string{stateNew}, Dst: stateOnProgress},
	{Name: EventComplete, Src: []string{stateOnProgress}, Dst: stateCompleted},
	// fail from new: the job could not be queued
	{Name: EventFail, Src: []string{stateNew, stateOnProgress}, Dst: stateFailed},
	{Name: EventCancel, Src: []string{stateNew, stateOnProgress}, Dst: stateCancelled},
	{Name: EventRetry, Src: []string{stateFailed, stateCancelled}, Dst: stateNew},
}

func newStateMachine(status int8) (*fsm.FSM, error) {
	state, ok := statusStates[status]
	if !ok {
		return nil, fmt.Errorf("unknown file management status %d", status)
	}

	return fsm.NewFSM(state, stateEvents, nil), nil
}

// nextStatus returns the status reached by firing event from status, or an error when the
// event is not allowed (fsm.InvalidEventError) or the status is unknown.
func nextStatus(ctx context.Context, status int8, event string) (int8, error) {
	stateMachine, err := newStateMachine(status)
	if err != nil {
		return 0, err
	}

	if err := stateMachine.Event(ctx, event); err != nil {
		return 0, err
	}

	return stateStatuses[stateMachine.Current()], nil
}

// availableEvents lists the events that can be fired from status, sorted.
func availableEvents(status int8) []string {
	stateMachine, err := newStateMachine(status)
	if err != nil {
		return []string{}
	}

	events := stateMachine.AvailableTransitions()
	sort.Strings(events)
	return events
}

func isTerminalStatus(status int8) bool {
	return status == constant.FileManagementCompleted ||
		status == constant.FileManagementFailed ||
		status == constant.FileManagementCancelled
}
