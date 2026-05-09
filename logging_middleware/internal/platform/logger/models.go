package logger

import (
	"errors"
	"fmt"
)

// LogLevel defines the allowed logging levels.
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

// LogStack defines the allowed stacks.
type LogStack string

const (
	StackBackend  LogStack = "backend"
	StackFrontend LogStack = "frontend"
)

// LogPackage defines the allowed packages.
type LogPackage string

const (
	PkgCache      LogPackage = "cache"
	PkgController LogPackage = "controller"
	PkgCronJob    LogPackage = "cron_job"
	PkgDB         LogPackage = "db"
	PkgDomain     LogPackage = "domain"
	PkgHandler    LogPackage = "handler"
	PkgRepository LogPackage = "repository"
	PkgRoute      LogPackage = "route"
	PkgService    LogPackage = "service"

	PkgAuth       LogPackage = "auth"
	PkgConfig     LogPackage = "config"
	PkgMiddleware LogPackage = "middleware"
	PkgUtils      LogPackage = "utils"
)

// LogPayload represents the structure sent to the evaluation service.
type LogPayload struct {
	Stack   string `json:"stack"`
	Level   string `json:"level"`
	Package string `json:"package"`
	Message string `json:"message"`
}

var (
	ErrInvalidStack   = errors.New("invalid stack")
	ErrInvalidLevel   = errors.New("invalid level")
	ErrInvalidPackage = errors.New("invalid package")
)

// Validate checks if the payload fields are within the allowed values.
func (p *LogPayload) Validate() error {
	switch LogStack(p.Stack) {
	case StackBackend, StackFrontend:
		// valid
	default:
		return fmt.Errorf("%w: %s", ErrInvalidStack, p.Stack)
	}

	switch LogLevel(p.Level) {
	case LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal:
		// valid
	default:
		return fmt.Errorf("%w: %s", ErrInvalidLevel, p.Level)
	}

	switch LogPackage(p.Package) {
	case PkgCache, PkgController, PkgCronJob, PkgDB, PkgDomain, PkgHandler,
		PkgRepository, PkgRoute, PkgService, PkgAuth, PkgConfig, PkgMiddleware, PkgUtils:
		// valid
	default:
		return fmt.Errorf("%w: %s", ErrInvalidPackage, p.Package)
	}

	return nil
}
