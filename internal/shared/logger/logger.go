// Package logger provides a structured zap logger configured per environment.
package logger

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"go.uber.org/zap"
)

// New builds a zap.Logger: JSON in production, human-friendly in development.
func New(cfg *config.Config) (*zap.Logger, error) {
	if cfg.IsProduction() {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

// Sugared returns the sugared variant.
func Sugared(l *zap.Logger) *zap.SugaredLogger { return l.Sugar() }
