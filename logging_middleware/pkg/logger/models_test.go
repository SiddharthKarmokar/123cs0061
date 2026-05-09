package logger

import (
	"errors"
	"testing"
)

func TestLogPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload LogPayload
		wantErr error
	}{
		{
			name: "Valid Backend Error",
			payload: LogPayload{
				Stack:   "backend",
				Level:   "error",
				Package: "handler",
				Message: "received string, expected bool",
			},
			wantErr: nil,
		},
		{
			name: "Valid Frontend Debug",
			payload: LogPayload{
				Stack:   "frontend",
				Level:   "debug",
				Package: "controller",
				Message: "test message",
			},
			wantErr: nil,
		},
		{
			name: "Invalid Stack",
			payload: LogPayload{
				Stack:   "mobile",
				Level:   "error",
				Package: "handler",
				Message: "test message",
			},
			wantErr: ErrInvalidStack,
		},
		{
			name: "Invalid Level",
			payload: LogPayload{
				Stack:   "backend",
				Level:   "critical",
				Package: "handler",
				Message: "test message",
			},
			wantErr: ErrInvalidLevel,
		},
		{
			name: "Invalid Package",
			payload: LogPayload{
				Stack:   "backend",
				Level:   "error",
				Package: "unknown_pkg",
				Message: "test message",
			},
			wantErr: ErrInvalidPackage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if tt.wantErr == nil && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
