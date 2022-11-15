package org

import (
	"fmt"
)

type Tenant struct {
	TenantID string `json:"tenant_id"`
	Audience string `json:"aud,omitempty"`
	Version  uint8  `json:"version"`
}

// Valid returns an error if JWT payload is incomplete
func (t *Tenant) Valid() error {
	if t.TenantID == "" {
		return fmt.Errorf("tenant is empty")
	}

	return nil
}
