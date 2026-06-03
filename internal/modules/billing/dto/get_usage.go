package dto

// UsageItem is one metered dimension's consumption vs. its quota.
type UsageItem struct {
	Kind      string `json:"kind"`
	Used      int64  `json:"used"`
	Quota     int64  `json:"quota"` // 0 = unlimited
	Unlimited bool   `json:"unlimited"`
}

// UsageResponse is the current-period usage rollup for the panel.
type UsageResponse struct {
	PeriodStart string      `json:"period_start"`
	PeriodEnd   string      `json:"period_end"`
	Items       []UsageItem `json:"items"`
}
