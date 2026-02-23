// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package model

import "fmt"

// ErrModelNotFound is returned when no scorer is registered for the requested model name.
type ErrModelNotFound struct {
	Name string
}

func (e *ErrModelNotFound) Error() string {
	return fmt.Sprintf("model not found: %s", e.Name)
}
