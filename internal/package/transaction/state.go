package transaction

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type State string

/*
const (

	StateCreated    State = "CREATED"
	StateStaged     State = "STAGED"
	StateCommitting State = "COMMITTING"
	StateCommitted  State = "COMMITTED"
	StateCompleted  State = "COMPLETED"

)
*/
const (
	StateCreated     State = "CREATED"
	StateStaged      State = "STAGED"
	StateCommitting  State = "COMMITTING"
	StateCommitted   State = "COMMITTED"
	StateFailed      State = "FAILED"
	StateRollingBack State = "ROLLING_BACK"
	StateRolledBack  State = "ROLLED_BACK"
)

type PersistentState struct {
	State State `json:"state"`
}

func statePath(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, "state")
}

func writeState(workspaceRoot string, state State) error {
	if workspaceRoot == "" {
		return fmt.Errorf("workspace root is empty")
	}

	data, err := json.MarshalIndent(
		PersistentState{State: state},
		"",
		"  ",
	)
	if err != nil {
		return fmt.Errorf("marshal transaction state: %w", err)
	}

	data = append(data, '\n')

	path := statePath(workspaceRoot)

	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write transaction state: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("commit transaction state: %w", err)
	}

	return nil
}

func readState(workspaceRoot string) (State, error) {
	if workspaceRoot == "" {
		return "", fmt.Errorf("workspace root is empty")
	}

	data, err := os.ReadFile(statePath(workspaceRoot))
	if err != nil {
		return "", fmt.Errorf("read transaction state: %w", err)
	}

	var persistent PersistentState

	if err := json.Unmarshal(data, &persistent); err != nil {
		return "", fmt.Errorf("invalid transaction state: %w", err)
	}

	return persistent.State, nil
}
