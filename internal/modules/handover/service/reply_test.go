package service

import (
	"context"
	"errors"
	"testing"

	convmodels "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- mocks for the dispatch dependencies -----------------------------------

type mockDispatchRepo struct {
	fn func(ctx context.Context, convID uuid.UUID) (*repository.DispatchInfo, error)
}

func (m mockDispatchRepo) LoadDispatchInfo(ctx context.Context, convID uuid.UUID) (*repository.DispatchInfo, error) {
	return m.fn(ctx, convID)
}

type mockRegistry struct {
	adapter channeladapter.Adapter
	ok      bool
}

func (m mockRegistry) For(string) (channeladapter.Adapter, bool) { return m.adapter, m.ok }

// mockCipher prefixes "dec:" so tests can assert the value was decrypted.
type mockCipher struct{}

func (mockCipher) Decrypt(enc string) (string, error) { return "dec:" + enc, nil }

// mockAdapter records the last SendText call and returns a canned id/err.
type mockAdapter struct {
	gotOut  channeladapter.Outbound
	gotText string
	msgID   string
	err     error
}

func (a *mockAdapter) Channel() string { return channeladapter.WhatsApp }
func (a *mockAdapter) SendText(_ context.Context, o channeladapter.Outbound, text string, _ int) (string, error) {
	a.gotOut, a.gotText = o, text
	return a.msgID, a.err
}
func (a *mockAdapter) SendPresence(context.Context, channeladapter.Outbound, string) error { return nil }
func (a *mockAdapter) MarkRead(context.Context, channeladapter.Outbound, string) error     { return nil }
func (a *mockAdapter) ConnectionState(context.Context, string) (string, error)             { return "", nil }

func humanConvRepo() *mockConvRepo {
	conv := &convmodels.Conversation{ID: uuid.New(), CompanyID: uuid.New(), State: convsvc.StateHuman}
	return &mockConvRepo{
		getConvFn: func(context.Context, uuid.UUID) (*convmodels.Conversation, error) { return conv, nil },
	}
}

func dispatchInfo() *repository.DispatchInfo {
	return &repository.DispatchInfo{
		Instance: "inst-1", APIKeyEnc: "rawkey", ChannelType: channeladapter.WhatsApp,
		RemoteJID: "5511988887777@s.whatsapp.net",
	}
}

// --- tests -----------------------------------------------------------------

func TestReply_LoadDispatchInfoError(t *testing.T) {
	boom := errors.New("db down")
	svc := &Service{
		conv: newConvSvc(humanConvRepo()), rdb: deadRedis(),
		repo: mockDispatchRepo{fn: func(context.Context, uuid.UUID) (*repository.DispatchInfo, error) {
			return nil, boom
		}},
		cipher: mockCipher{}, adapters: mockRegistry{ok: true},
	}
	require.ErrorIs(t, svc.Reply(context.Background(), uuid.New(), "hi"), boom)
}

func TestReply_NoAdapterForChannel(t *testing.T) {
	svc := &Service{
		conv: newConvSvc(humanConvRepo()), rdb: deadRedis(),
		repo:     mockDispatchRepo{fn: func(context.Context, uuid.UUID) (*repository.DispatchInfo, error) { return dispatchInfo(), nil }},
		cipher:   mockCipher{},
		adapters: mockRegistry{ok: false}, // unknown channel
	}
	require.ErrorIs(t, svc.Reply(context.Background(), uuid.New(), "hi"), ErrNoChannel)
}

func TestReply_SendTextError(t *testing.T) {
	boom := errors.New("evolution 500")
	svc := &Service{
		conv: newConvSvc(humanConvRepo()), rdb: deadRedis(),
		repo:     mockDispatchRepo{fn: func(context.Context, uuid.UUID) (*repository.DispatchInfo, error) { return dispatchInfo(), nil }},
		cipher:   mockCipher{},
		adapters: mockRegistry{adapter: &mockAdapter{err: boom}, ok: true},
	}
	require.ErrorIs(t, svc.Reply(context.Background(), uuid.New(), "hi"), boom)
}

func TestReply_Success_DecryptsAndStripsJID(t *testing.T) {
	ad := &mockAdapter{msgID: "wamid.123"}
	svc := &Service{
		conv: newConvSvc(humanConvRepo()), rdb: deadRedis(),
		repo:     mockDispatchRepo{fn: func(context.Context, uuid.UUID) (*repository.DispatchInfo, error) { return dispatchInfo(), nil }},
		cipher:   mockCipher{},
		adapters: mockRegistry{adapter: ad, ok: true},
	}
	require.NoError(t, svc.Reply(context.Background(), uuid.New(), "hello operator"))
	require.Equal(t, "hello operator", ad.gotText)
	require.Equal(t, "dec:rawkey", ad.gotOut.APIKey, "api key must be decrypted before dispatch")
	require.Equal(t, "5511988887777", ad.gotOut.Number, "remote JID suffix must be stripped")
	require.Equal(t, "inst-1", ad.gotOut.Instance)
}
