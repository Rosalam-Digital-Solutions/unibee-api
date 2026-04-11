package user

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"unibee/api/user/plan"
	"unibee/internal/cmd/config"
	"unibee/internal/consts"
	_interface "unibee/internal/interface/context"
	plan2 "unibee/internal/logic/plan"
	"unibee/internal/query"
	"unibee/utility"
)

func (c *ControllerPlan) List(ctx context.Context, req *plan.ListReq) (res *plan.ListRes, err error) {
	if !config.GetConfigInstance().IsLocal() {
		utility.Assert(_interface.Context().Get(ctx).User != nil, "auth failure,not login")
	}

	if len(req.ProductIds) == 0 {
		req.ProductIds = append(req.ProductIds, 0)
	}

	publishPlans, total := plan2.PlanList(ctx, &plan2.ListInternalReq{
		MerchantId:    _interface.GetMerchantId(ctx),
		ProductIds:    req.ProductIds,
		Type:          req.Type,
		Status:        []int{consts.PlanStatusActive},
		PublishStatus: consts.PlanPublishStatusPublished,
		Currency:      req.Currency,
		SearchKey:     req.SearchKey,
		Page:          req.Page,
		Count:         req.Count,
	})
	for _, productId := range req.ProductIds {
		sub := query.GetLatestActiveOrIncompleteOrCreateSubscriptionByUserId(ctx, _interface.Context().Get(ctx).User.Id, _interface.GetMerchantId(ctx), productId)
		if sub != nil {
			subPlan := query.GetPlanById(ctx, sub.PlanId)
			if subPlan != nil && subPlan.PublishStatus != consts.PlanPublishStatusPublished {
				userPlan, _ := plan2.PlanDetail(ctx, _interface.GetMerchantId(ctx), subPlan.Id)
				if userPlan != nil {
					publishPlans = append(publishPlans, userPlan.Plan)
				}
			}
		}
	}
	g.Log().Infof(ctx, "UserPlanList merchantId:%d userId:%d productIds:%v type:%v returnedPlans:%d total:%d",
		_interface.GetMerchantId(ctx),
		_interface.Context().Get(ctx).User.Id,
		req.ProductIds,
		req.Type,
		len(publishPlans),
		total,
	)
	return &plan.ListRes{Plans: publishPlans, Total: total}, nil
}
