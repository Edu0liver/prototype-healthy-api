package http

import (
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// GetUsage handles GET /billing/usage.
// @Summary  Get current-period usage
// @Tags     billing
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} dto.UsageResponse
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Router   /billing/usage [get]
func (h *Handler) GetUsage(c *gin.Context) {
	companyID := appctx.CompanyID(c.Request.Context())
	sum, err := h.svc.GetUsage(c.Request.Context(), companyID)
	if err != nil {
		httputil.Fail(c, err)
		return
	}

	quotas := map[string]int64{
		models.KindAIMessage:    int64(sum.Limits.QuotaAIMessages),
		models.KindLLMTokens:    sum.Limits.QuotaTokens,
		models.KindAudioMinutes: int64(sum.Limits.QuotaAudioMinutes),
		models.KindStorageMB:    int64(sum.Limits.QuotaStorageMB),
	}
	items := make([]dto.UsageItem, 0, len(quotas))
	for _, kind := range []string{models.KindAIMessage, models.KindLLMTokens, models.KindAudioMinutes, models.KindStorageMB} {
		q := quotas[kind]
		items = append(items, dto.UsageItem{
			Kind:      kind,
			Used:      sum.Totals[kind],
			Quota:     q,
			Unlimited: q == 0,
		})
	}

	httputil.OK(c, dto.UsageResponse{
		PeriodStart: sum.PeriodStart.Format(time.RFC3339),
		PeriodEnd:   sum.PeriodEnd.Format(time.RFC3339),
		Items:       items,
	})
}
