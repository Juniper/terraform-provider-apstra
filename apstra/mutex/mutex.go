package mutex

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"encoding/json"
	"fmt"
	"terraform-provider-apstra/apstra/utils"
	"time"
)

const (
	retryInterval = 500 * time.Millisecond
)

type BlueprintMutexMsg struct {
	Owner   string `json:"owner"`
	Details string `json:"details"`
}

func (o *BlueprintMutexMsg) String() (string, error) {
	msgData, err := json.Marshal(o)
	if err != nil {
		return "", fmt.Errorf("error marshaling blueprint mutex message - %w", err)
	}
	return string(msgData), nil
}

func sameOwner(a, b *goapstra.TwoStageL3ClosMutex) bool {
	msgA := &BlueprintMutexMsg{}
	msgB := &BlueprintMutexMsg{}

	var err error

	// we don't return unmarshal errors here because, while we can't assume the
	// field contains valid JSON, we *can* conclude that "owner" field won't
	// match if it's not valid JSON.
	err = json.Unmarshal([]byte(a.GetMessage()), msgA)
	if err != nil {
		return false
	}

	err = json.Unmarshal([]byte(b.GetMessage()), msgB)
	if err != nil {
		return false
	}

	return msgA.Owner == msgB.Owner
}

func Lock(ctx context.Context, mutex *goapstra.TwoStageL3ClosMutex) error {
	ticker := utils.ImmediateTicker(retryInterval)
	defer ticker.Stop()
	for {
		// mind the timeout
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		// attempt to lock
		ok, reason, err := mutex.TryLock(ctx, true)
		if err != nil {
			// oops
			return fmt.Errorf("error locking blueprint mutex - %w", err)
		}
		if ok {
			// we got the lock
			return nil
		}
		if reason != nil && sameOwner(mutex, reason) {
			// safe to proceed because lock has same owner
			return nil
		}
	}
}
