package file_managements

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/looplab/fsm"

	"go-api/constant"
)

func TestNextStatusAllowedTransitions(t *testing.T) {
	cases := []struct {
		from  int8
		event string
		to    int8
	}{
		{constant.FileManagementNew, EventProcess, constant.FileManagementOnProgress},
		{constant.FileManagementNew, EventCancel, constant.FileManagementCancelled},
		{constant.FileManagementNew, EventFail, constant.FileManagementFailed},
		{constant.FileManagementOnProgress, EventComplete, constant.FileManagementCompleted},
		{constant.FileManagementOnProgress, EventFail, constant.FileManagementFailed},
		{constant.FileManagementOnProgress, EventCancel, constant.FileManagementCancelled},
		{constant.FileManagementFailed, EventRetry, constant.FileManagementNew},
		{constant.FileManagementCancelled, EventRetry, constant.FileManagementNew},
	}

	for _, c := range cases {
		to, err := nextStatus(context.Background(), c.from, c.event)
		if err != nil || to != c.to {
			t.Errorf("%d --%s--> got %d (%v), want %d", c.from, c.event, to, err, c.to)
		}
	}
}

func TestNextStatusRejectsInvalidEvents(t *testing.T) {
	cases := []struct {
		from  int8
		event string
	}{
		{constant.FileManagementNew, EventComplete},
		{constant.FileManagementNew, EventRetry},
		{constant.FileManagementOnProgress, EventProcess},
		{constant.FileManagementCompleted, EventCancel},
		{constant.FileManagementCompleted, EventRetry},
		{constant.FileManagementFailed, EventCancel},
		{constant.FileManagementCancelled, EventProcess},
	}

	for _, c := range cases {
		_, err := nextStatus(context.Background(), c.from, c.event)
		var invalidEvent fsm.InvalidEventError
		if !errors.As(err, &invalidEvent) {
			t.Errorf("%d --%s--> err = %v, want InvalidEventError", c.from, c.event, err)
		}
	}

	if _, err := nextStatus(context.Background(), 42, EventProcess); err == nil {
		t.Error("unknown status should fail")
	}
}

func TestUserEvents(t *testing.T) {
	cases := map[int8][]string{
		constant.FileManagementNew:        {EventCancel},
		constant.FileManagementOnProgress: {EventCancel},
		constant.FileManagementFailed:     {EventRetry},
		constant.FileManagementCancelled:  {EventRetry},
		constant.FileManagementCompleted:  {},
	}

	for status, want := range cases {
		if got := userEvents(status); !reflect.DeepEqual(got, want) {
			t.Errorf("userEvents(%d) = %v, want %v", status, got, want)
		}
	}
}
