package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unibee/internal/consts"
	"unibee/internal/logic/gateway/api"
	"unibee/internal/logic/gateway/api/log"
	"unibee/internal/logic/gateway/gateway_bean"
	"unibee/internal/logic/gateway/util"
	handler2 "unibee/internal/logic/payment/handler"
	entity "unibee/internal/model/entity/default"
	"unibee/internal/query"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type ChapaWebhook struct {
}

func (c ChapaWebhook) GatewayNewPaymentMethodRedirect(r *ghttp.Request, gateway *entity.MerchantGateway) (err error) {
	return nil
}

func (c ChapaWebhook) GatewayCheckAndSetupWebhook(ctx context.Context, gateway *entity.MerchantGateway) (err error) {
	secret := strings.TrimSpace(gateway.GatewaySecret)
	if len(secret) == 0 {
		secret = strings.TrimSpace(gateway.GatewayKey)
	}
	_ = query.UpdateGatewayWebhookSecret(ctx, gateway.Id, secret)
	return nil
}

func (c ChapaWebhook) GatewayWebhook(r *ghttp.Request, gateway *entity.MerchantGateway) {
	if strings.EqualFold(r.Method, "GET") {
		// Chapa webhooks are POST requests. Browsers often trigger GET when the URL
		// is opened manually; return 200 to avoid false integration errors.
		r.Response.WriteStatus(200)
		r.Response.Writeln("ok")
		return
	}

	body := r.GetBody()
	jsonData, err := gjson.LoadJson(string(body))
	if err != nil {
		g.Log().Errorf(r.Context(), "Webhook Gateway:%s, parse body failed: %s", gateway.GatewayName, err.Error())
		r.Response.WriteStatusExit(400)
		return
	}

	if !verifyChapaSignature(r, gateway, body) {
		log.SaveChannelHttpLog("GatewayWebhook", string(body), "SignatureVerifyError-400", nil, fmt.Sprintf("%s-%d", gateway.GatewayName, gateway.Id), nil, gateway)
		r.Response.WriteStatusExit(400)
		return
	}

	metaPaymentId := jsonData.Get("tx_ref").String()
	if len(metaPaymentId) == 0 {
		metaPaymentId = jsonData.Get("trx_ref").String()
	}
	gatewayPaymentId := jsonData.Get("chapa_reference").String()
	if len(gatewayPaymentId) == 0 {
		gatewayPaymentId = jsonData.Get("reference").String()
	}
	if len(gatewayPaymentId) == 0 {
		gatewayPaymentId = metaPaymentId
	}

	if len(metaPaymentId) == 0 {
		log.SaveChannelHttpLog("GatewayWebhook", jsonData, "Missing tx_ref", nil, fmt.Sprintf("%s-%d", gateway.GatewayName, gateway.Id), nil, gateway)
		r.Response.WriteStatusExit(400)
		return
	}

	err = ProcessPaymentWebhook(r.Context(), metaPaymentId, gatewayPaymentId, gateway)
	if err != nil {
		g.Log().Errorf(r.Context(), "Webhook Gateway:%s, ProcessPaymentWebhook Error:%s", gateway.GatewayName, err.Error())
		log.SaveChannelHttpLog("GatewayWebhook", jsonData, err.Error(), err, fmt.Sprintf("%s-%d", gateway.GatewayName, gateway.Id), nil, gateway)
		r.Response.WriteStatusExit(400)
		return
	}

	log.SaveChannelHttpLog("GatewayWebhook", jsonData, "ok", nil, fmt.Sprintf("%s-%d", gateway.GatewayName, gateway.Id), nil, gateway)
	r.Response.WriteStatus(200)
}

func (c ChapaWebhook) GatewayRedirect(r *ghttp.Request, gateway *entity.MerchantGateway) (res *gateway_bean.GatewayRedirectResp, err error) {
	payIdStr := r.Get("paymentId").String()
	if len(payIdStr) == 0 {
		payIdStr = r.Get("tx_ref").String()
	}
	if len(payIdStr) == 0 {
		payIdStr = r.Get("trx_ref").String()
	}

	response := ""
	status := false
	returnUrl := ""
	isSuccess := false
	var payment *entity.Payment

	if len(payIdStr) > 0 {
		payment = query.GetPaymentByPaymentId(r.Context(), payIdStr)
		if payment != nil {
			successRaw := r.Get("success").String()
			if len(successRaw) == 0 {
				successRaw = r.Get("status").String()
			}
			if strings.EqualFold(successRaw, "true") || strings.EqualFold(successRaw, "success") {
				isSuccess = true
				successRaw = "true"
			} else {
				successRaw = "false"
			}
			returnUrl = util.GetPaymentRedirectUrl(r.Context(), payment, successRaw)
		}

		if isSuccess {
			if payment == nil || len(payment.GatewayPaymentIntentId) == 0 {
				response = "paymentId invalid"
			} else if len(payment.GatewayPaymentId) > 0 && payment.Status == consts.PaymentSuccess {
				response = "success"
				status = true
			} else {
				paymentIntentDetail, detailErr := api.GetGatewayServiceProvider(r.Context(), gateway.Id).GatewayPaymentDetail(r.Context(), gateway, payment.GatewayPaymentId, payment)
				if detailErr != nil {
					response = fmt.Sprintf("%v", detailErr)
				} else if paymentIntentDetail.Status == consts.PaymentSuccess {
					handleErr := handler2.HandlePaySuccess(r.Context(), &handler2.HandlePayReq{
						PaymentId:              payment.PaymentId,
						GatewayPaymentIntentId: payment.GatewayPaymentIntentId,
						GatewayPaymentId:       paymentIntentDetail.GatewayPaymentId,
						GatewayUserId:          paymentIntentDetail.GatewayUserId,
						TotalAmount:            paymentIntentDetail.TotalAmount,
						PayStatusEnum:          consts.PaymentSuccess,
						PaidTime:               paymentIntentDetail.PaidTime,
						PaymentAmount:          paymentIntentDetail.PaymentAmount,
						Reason:                 paymentIntentDetail.Reason,
						GatewayPaymentMethod:   paymentIntentDetail.GatewayPaymentMethod,
					})
					if handleErr != nil {
						response = fmt.Sprintf("%v", handleErr)
					} else {
						response = "payment success"
						status = true
					}
				} else if paymentIntentDetail.Status == consts.PaymentFailed || paymentIntentDetail.Status == consts.PaymentCancelled {
					handleErr := handler2.HandlePayFailure(r.Context(), &handler2.HandlePayReq{
						PaymentId:              payment.PaymentId,
						GatewayPaymentIntentId: payment.GatewayPaymentIntentId,
						GatewayPaymentId:       paymentIntentDetail.GatewayPaymentId,
						PayStatusEnum:          consts.PaymentStatusEnum(paymentIntentDetail.Status),
						Reason:                 paymentIntentDetail.Reason,
					})
					if handleErr != nil {
						response = fmt.Sprintf("%v", handleErr)
					}
				}
			}
		} else {
			response = "user cancelled"
		}
	}

	log.SaveChannelHttpLog("GatewayRedirect", r.URL, response, err, fmt.Sprintf("%s-%d", gateway.GatewayName, gateway.Id), nil, gateway)
	return &gateway_bean.GatewayRedirectResp{
		Payment:   payment,
		Status:    status,
		Message:   response,
		Success:   isSuccess,
		ReturnUrl: returnUrl,
		QueryPath: r.URL.RawQuery,
	}, nil
}

func verifyChapaSignature(r *ghttp.Request, gateway *entity.MerchantGateway, body []byte) bool {
	provided := strings.TrimSpace(r.Header.Get("x-chapa-signature"))
	if len(provided) == 0 {
		provided = strings.TrimSpace(r.Header.Get("chapa-signature"))
	}
	if len(provided) == 0 {
		return true
	}

	secret := strings.TrimSpace(gateway.WebhookSecret)
	if len(secret) == 0 {
		secret = strings.TrimSpace(gateway.GatewaySecret)
	}
	if len(secret) == 0 {
		secret = strings.TrimSpace(gateway.GatewayKey)
	}
	if len(secret) == 0 {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if strings.EqualFold(expected, provided) {
		return true
	}

	macLegacy := hmac.New(sha256.New, []byte(secret))
	_, _ = macLegacy.Write([]byte(secret))
	expectedLegacy := hex.EncodeToString(macLegacy.Sum(nil))
	return strings.EqualFold(expectedLegacy, provided)
}
