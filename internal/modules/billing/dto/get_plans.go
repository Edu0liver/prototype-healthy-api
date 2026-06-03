package dto

// PlanResponse is one entry of the public plan catalogue.
type PlanResponse struct {
	Code              string `json:"code"`
	Name              string `json:"name"`
	PriceCents        int    `json:"price_cents"`
	Currency          string `json:"currency"`
	QuotaAIMessages   int    `json:"quota_ai_messages"`
	QuotaTokens       int64  `json:"quota_tokens"`
	QuotaAudioMinutes int    `json:"quota_audio_minutes"`
	QuotaStorageMB    int    `json:"quota_storage_mb"`
	MaxChannels       int    `json:"max_channels"`
	MaxAgents         int    `json:"max_agents"`
	MaxKB             int    `json:"max_kb"`
	MaxSeats          int    `json:"max_seats"`
	Purchasable       bool   `json:"purchasable"` // has a Stripe price configured
}

// PlansResponse wraps the plan catalogue.
type PlansResponse struct {
	Plans []PlanResponse `json:"plans"`
}
