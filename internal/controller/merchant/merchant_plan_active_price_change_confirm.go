package merchant

import (
	"context"
	"unibee/api/bean"
	"unibee/api/merchant/plan"
	_interface "unibee/internal/interface/context"
	plan2 "unibee/internal/logic/plan"
)

func (c *ControllerPlan) ActivePriceChangeConfirm(ctx context.Context, req *plan.ActivePriceChangeConfirmReq) (res *plan.ActivePriceChangeConfirmRes, err error) {
	one, err := plan2.ActivePlanPriceChangeConfirm(ctx, _interface.GetMerchantId(ctx), req.PlanId, req.NewAmount, req.ConfirmOldAmount, req.Reason)
	if err != nil {
		return nil, err
	}
	return &plan.ActivePriceChangeConfirmRes{
		Plan: bean.SimplifyPlanWithContext(ctx, one),
	}, nil
}
