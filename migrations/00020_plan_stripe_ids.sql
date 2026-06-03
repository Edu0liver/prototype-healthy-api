-- +goose Up
-- Map each plan to its Stripe price/product so checkout and subscription
-- webhooks can translate between Stripe objects and local plans.
ALTER TABLE plans ADD COLUMN stripe_price_id   text;
ALTER TABLE plans ADD COLUMN stripe_product_id text;

CREATE UNIQUE INDEX uniq_plans_stripe_price
    ON plans (stripe_price_id)
    WHERE stripe_price_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS uniq_plans_stripe_price;
ALTER TABLE plans DROP COLUMN IF EXISTS stripe_product_id;
ALTER TABLE plans DROP COLUMN IF EXISTS stripe_price_id;
