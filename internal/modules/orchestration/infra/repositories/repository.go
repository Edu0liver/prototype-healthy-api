// Package repositories provides the read models the orchestration worker needs:
// channel credentials and the active automation's agent configuration. It reads
// other modules' tables directly (tenant-scoped) to avoid importing them.
package repositories

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/platform/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNoActiveAgent indicates the channel has no active automation/agent.
var ErrNoActiveAgent = errors.New("orchestration: no active agent for channel")

// Repository reads orchestration inputs.
type Repository struct{}

// New builds the repository.
func New() *Repository { return &Repository{} }

// ChannelCreds holds the data needed to talk to a channel's provider.
type ChannelCreds struct {
	InstanceName string `gorm:"column:evolution_instance_name"`
	APIKeyEnc    string `gorm:"column:evolution_apikey_enc"`
	Type         string `gorm:"column:type"`
}

// LoadChannelCreds loads instance name + encrypted apikey for a channel.
func (r *Repository) LoadChannelCreds(ctx context.Context, channelID uuid.UUID) (*ChannelCreds, error) {
	var c ChannelCreds
	err := database.MustTx(ctx).Table("channels").Scopes(database.TenantScope(ctx)).
		Select("evolution_instance_name, evolution_apikey_enc, type").
		Where("id = ?", channelID).Take(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoActiveAgent
	}
	return &c, err
}

// AgentConfig is the active agent + automation rules for a channel.
type AgentConfig struct {
	AgentID          uuid.UUID                `gorm:"column:agent_id"`
	SystemPrompt     string                   `gorm:"column:system_prompt"`
	Model            string                   `gorm:"column:model"`
	Temperature      float64                  `gorm:"column:temperature"`
	MaxOutputTokens  int                      `gorm:"column:max_output_tokens"`
	HandoverEnabled  bool                     `gorm:"column:handover_enabled"`
	HandoverKeywords database.JSONStringArray `gorm:"column:handover_keywords"`
	FallbackMessage  string                   `gorm:"column:fallback_message"`
	DebounceSeconds  int                      `gorm:"column:debounce_seconds"`
}

// LoadActiveAgent returns the active automation's agent config for a channel.
func (r *Repository) LoadActiveAgent(ctx context.Context, channelID uuid.UUID) (*AgentConfig, error) {
	var cfg AgentConfig
	// Explicit au.company_id (not TenantScope) to avoid ambiguous company_id
	// across the joined tables.
	err := database.MustTx(ctx).
		Table("automations AS au").
		Select("a.id AS agent_id, a.system_prompt, a.model, a.temperature, a.max_output_tokens, a.handover_enabled, a.handover_keywords, au.fallback_message, au.debounce_seconds").
		Joins("JOIN agents a ON a.id = au.agent_id").
		Where("au.company_id = ? AND au.channel_id = ? AND au.is_active = ?", appctx.CompanyID(ctx), channelID, true).
		Take(&cfg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoActiveAgent
	}
	return &cfg, err
}
