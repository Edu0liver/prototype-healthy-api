package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// GetPlans handles GET /billing/plans.
// @Summary  List the plan catalogue
// @Tags     billing
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} dto.PlansResponse
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Router   /billing/plans [get]
func (h *Handler) GetPlans(c *gin.Context) {
	plans, err := h.svc.ListPlans(c.Request.Context())
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dto.PlanResponse, 0, len(plans))
	for _, p := range plans {
		out = append(out, dto.PlanResponse{
			Code:              p.Code,
			Name:              p.Name,
			PriceCents:        p.PriceCents,
			Currency:          p.Currency,
			QuotaAIMessages:   p.QuotaAIMessages,
			QuotaTokens:       p.QuotaTokens,
			QuotaAudioMinutes: p.QuotaAudioMinutes,
			QuotaStorageMB:    p.QuotaStorageMB,
			MaxChannels:       p.MaxChannels,
			MaxAgents:         p.MaxAgents,
			MaxKB:             p.MaxKB,
			MaxSeats:          p.MaxSeats,
			Purchasable:       p.StripePriceID != "",
		})
	}
	httputil.OK(c, dto.PlansResponse{Plans: out})
}
