-- +goose Up
-- Drop the Free tier: the platform minimum is now R$14,99 (starter). Existing
-- free subscriptions are repointed to starter and marked `canceled`, so those
-- tenants hold no valid plan and are blocked until they subscribe (enforced by
-- the billing quota checks + subscription gate).
-- +goose StatementBegin
DO $$
DECLARE
    starter_id uuid;
    free_id    uuid;
BEGIN
    SELECT id INTO starter_id FROM plans WHERE code = 'starter';
    SELECT id INTO free_id FROM plans WHERE code = 'free';
    IF free_id IS NOT NULL THEN
        UPDATE subscriptions
            SET plan_id = starter_id, status = 'canceled', updated_at = now()
            WHERE plan_id = free_id;
        DELETE FROM plans WHERE id = free_id;
    END IF;
END $$;
-- +goose StatementEnd

-- Restructure the paid tiers. Starter is the new entry point at R$14,99.
UPDATE plans SET
    price_cents = 1499, name = 'Starter',
    quota_ai_messages = 1000, quota_tokens = 2000000,
    quota_audio_minutes = 60, quota_storage_mb = 200,
    max_channels = 1, max_agents = 2, max_kb = 3, max_seats = 2,
    updated_at = now()
WHERE code = 'starter';

UPDATE plans SET
    price_cents = 9990, name = 'Pro',
    quota_ai_messages = 10000, quota_tokens = 20000000,
    quota_audio_minutes = 600, quota_storage_mb = 2000,
    max_channels = 5, max_agents = 10, max_kb = 20, max_seats = 10,
    updated_at = now()
WHERE code = 'pro';

-- Enterprise stays price 0 = "sob consulta" (custom, not self-serve checkout:
-- stripe_price_id is NULL so /billing/checkout returns 409).

-- Adopt the recommended RAG model (gpt-4.1-mini) for agents still on the old
-- default. Runs as superuser (bypasses RLS).
UPDATE agents SET model = 'gpt-4.1-mini', updated_at = now()
WHERE model = 'gpt-4o-mini';

-- companies.plan is a denormalized cache; clear the now-invalid 'free' value.
UPDATE companies SET plan = '', updated_at = now() WHERE plan = 'free';

-- +goose Down
-- Best-effort: recreate the Free plan (canceled subscriptions are not restored).
-- +goose StatementBegin
INSERT INTO plans
    (code, name, price_cents, currency,
     quota_ai_messages, quota_tokens, quota_audio_minutes, quota_storage_mb,
     max_channels, max_agents, max_kb, max_seats,
     overage_per_msg_cents, overage_per_1k_tokens_cents)
VALUES ('free', 'Free', 0, 'BRL', 100, 50000, 10, 50, 1, 1, 1, 1, 0, 0)
ON CONFLICT (code) DO NOTHING;
-- +goose StatementEnd
