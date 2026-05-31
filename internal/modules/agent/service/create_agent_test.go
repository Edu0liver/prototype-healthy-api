package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreate_DefaultsAndTenant(t *testing.T) {
	company := uuid.New()
	ctx := appctx.With(context.Background(), appctx.Identity{CompanyID: company})

	var saved *models.Agent
	svc := New(&mockRepo{createFn: func(_ context.Context, a *models.Agent) error {
		saved = a
		return nil
	}})

	out, err := svc.Create(ctx, dto.CreateAgentRequest{Name: "Bot", SystemPrompt: "hi"})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, company, out.CompanyID, "company must come from ctx identity")
	require.NotEqual(t, uuid.Nil, out.ID, "id must be generated")
	// Defaults applied when request omits optional fields.
	require.Equal(t, "gpt-4o-mini", out.Model)
	require.Equal(t, 0.7, out.Temperature)
	require.Equal(t, 1024, out.MaxOutputTokens)
	require.True(t, out.HandoverEnabled)
	require.Equal(t, "draft", out.Status)
	require.Same(t, saved, out, "returned agent must be the persisted one")
}

func TestCreate_OverridesDefaults(t *testing.T) {
	ctx := appctx.With(context.Background(), appctx.Identity{CompanyID: uuid.New()})
	temp := 0.1
	maxTok := 256
	handover := false
	svc := New(&mockRepo{})

	out, err := svc.Create(ctx, dto.CreateAgentRequest{
		Name:             "Bot",
		SystemPrompt:     "hi",
		Model:            "gpt-4o",
		Temperature:      &temp,
		MaxOutputTokens:  &maxTok,
		HandoverEnabled:  &handover,
		HandoverKeywords: []string{"human", "agent"},
		Status:           "active",
	})
	require.NoError(t, err)
	require.Equal(t, "gpt-4o", out.Model)
	require.Equal(t, 0.1, out.Temperature)
	require.Equal(t, 256, out.MaxOutputTokens)
	require.False(t, out.HandoverEnabled)
	require.Equal(t, []string{"human", "agent"}, []string(out.HandoverKeywords))
	require.Equal(t, "active", out.Status)
}

func TestCreate_RepoError(t *testing.T) {
	ctx := appctx.With(context.Background(), appctx.Identity{CompanyID: uuid.New()})
	boom := errors.New("db down")
	svc := New(&mockRepo{createFn: func(context.Context, *models.Agent) error { return boom }})

	out, err := svc.Create(ctx, dto.CreateAgentRequest{Name: "Bot", SystemPrompt: "hi"})
	require.Nil(t, out)
	require.ErrorIs(t, err, boom)
}
