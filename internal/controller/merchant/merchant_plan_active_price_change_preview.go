package merchant

import (
	"context"
	"unibee/api/merchant/plan"
	_interface "unibee/internal/interface/context"
	plan2 "unibee/internal/logic/plan"
)

func (c *ControllerPlan) ActivePriceChangePreview(ctx context.Context, req *plan.ActivePriceChangePreviewReq) (res *plan.ActivePriceChangePreviewRes, err error) {
	preview, err := plan2.ActivePlanPriceChangePreview(ctx, _interface.GetMerchantId(ctx), req.PlanId, req.NewAmount)
	if err != nil {
		return nil, err
	}
	return &plan.ActivePriceChangePreviewRes{
		ActiveAffectedSubscriptions: preview.ActiveAffectedSubscriptions,
		TotalAffectedSubscriptions:  preview.TotalAffectedSubscriptions,
	}, nil
}
