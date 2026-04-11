package api

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unibee/internal/consts"
	_interface "unibee/internal/interface"
	webhook2 "unibee/internal/logic/gateway"
	"unibee/internal/logic/gateway/api/log"
	"unibee/internal/logic/gateway/gateway_bean"
	entity "unibee/internal/model/entity/default"
	"unibee/utility"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"
)

const (
	chapaAPIBaseURL = "https://api.chapa.co"
)

var chapaDescriptionSanitizer = regexp.MustCompile(`[^A-Za-z0-9._\-\s]`)

var chapaSupportedCurrencies = map[string]bool{
	"ETB": true,
	"USD": true,
}

type Chapa struct {
}

func (c Chapa) GatewayInfo(ctx context.Context) *_interface.GatewayInfo {
	return &_interface.GatewayInfo{
		Name:                          "Chapa",
		Description:                   "Use Chapa secret key to initialize and verify checkout payments.",
		DisplayName:                   "Chapa",
		GatewayWebsiteLink:            "https://chapa.co",
		GatewayWebhookIntegrationLink: "https://developer.chapa.co/integrations/webhooks",
		GatewayLogo:                   "https://chapa.co/wp-content/uploads/2022/09/chapa-logo-dark.svg",
		GatewayIcons:                  []string{"https://chapa.co/wp-content/uploads/2022/09/chapa-logo-dark.svg"},
		GatewayType:                   consts.GatewayTypeCard,
		Sort:                          68,
		PublicKeyName:                 "Chapa Secret Key",
		PrivateSecretName:             "Webhook Secret (optional)",
	}
}

func (c Chapa) GatewayCryptoFiatTrans(ctx context.Context, from *gateway_bean.GatewayCryptoFromCurrencyAmountDetailReq) (to *gateway_bean.GatewayCryptoToCurrencyAmountDetailRes, err error) {
	return nil, gerror.New("not support")
}

func (c Chapa) GatewayTest(ctx context.Context, req *_interface.GatewayTestReq) (icon string, gatewayType int64, err error) {
	apiKey := c.resolveApiKey(req.Key, req.Secret)
	utility.Assert(len(apiKey) > 0, "Chapa secret key is required")

	payload := map[string]interface{}{
		"amount":       "10",
		"currency":     "ETB",
		"email":        "test@unibee.dev",
		"first_name":   "UniBee",
		"last_name":    "Test",
		"tx_ref":       fmt.Sprintf("ub-chapa-test-%s", uuid.New().String()),
		"callback_url": "https://example.com/chapa/callback",
		"return_url":   "https://example.com/chapa/return",
	}

	responseJson, err := sendChapaRequest(ctx, apiKey, "POST", "/v1/transaction/initialize", payload)
	utility.Assert(err == nil, fmt.Sprintf("invalid key, call error %s", err))
	utility.Assert(responseJson != nil && responseJson.Contains("data.checkout_url"), "invalid key, checkout_url is nil")

	return "https://chapa.co/wp-content/uploads/2022/09/chapa-logo-dark.svg", consts.GatewayTypeCard, nil
}

func (c Chapa) GatewayUserCreate(ctx context.Context, gateway *entity.MerchantGateway, user *entity.UserAccount) (res *gateway_bean.GatewayUserCreateResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayUserDetailQuery(ctx context.Context, gateway *entity.MerchantGateway, gatewayUserId string) (res *gateway_bean.GatewayUserDetailQueryResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayMerchantBalancesQuery(ctx context.Context, gateway *entity.MerchantGateway) (res *gateway_bean.GatewayMerchantBalanceQueryResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayUserAttachPaymentMethodQuery(ctx context.Context, gateway *entity.MerchantGateway, userId uint64, gatewayPaymentMethod string) (res *gateway_bean.GatewayUserAttachPaymentMethodResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayUserDeAttachPaymentMethodQuery(ctx context.Context, gateway *entity.MerchantGateway, userId uint64, gatewayPaymentMethod string) (res *gateway_bean.GatewayUserDeAttachPaymentMethodResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayUserPaymentMethodListQuery(ctx context.Context, gateway *entity.MerchantGateway, req *gateway_bean.GatewayUserPaymentMethodReq) (res *gateway_bean.GatewayUserPaymentMethodListResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayUserCreateAndBindPaymentMethod(ctx context.Context, gateway *entity.MerchantGateway, userId uint64, currency string, metadata map[string]interface{}) (res *gateway_bean.GatewayUserPaymentMethodCreateAndBindResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayNewPayment(ctx context.Context, gateway *entity.MerchantGateway, createPayContext *gateway_bean.GatewayNewPaymentReq) (res *gateway_bean.GatewayNewPaymentResp, err error) {
	apiKey := c.resolveApiKey(gateway.GatewayKey, gateway.GatewaySecret)
	utility.Assert(len(apiKey) > 0, "Chapa secret key is required")

	name, description := createPayContext.GetInvoiceSingleProductNameAndDescription()
	if len(name) == 0 {
		name = "UniBee Checkout"
	}
	if len(description) == 0 {
		description = name
	}
	name = sanitizeChapaText(name)
	description = sanitizeChapaText(description)

	currency := strings.ToUpper(createPayContext.Pay.Currency)
	totalAmount := createPayContext.Pay.TotalAmount
	if createPayContext.GatewayCurrencyExchange != nil && createPayContext.ExchangeAmount > 0 && len(createPayContext.ExchangeCurrency) > 0 {
		currency = strings.ToUpper(createPayContext.ExchangeCurrency)
		totalAmount = createPayContext.ExchangeAmount
	}
	if len(strings.TrimSpace(gateway.Currency)) > 0 {
		currency = strings.ToUpper(strings.TrimSpace(gateway.Currency))
	}
	utility.Assert(chapaSupportedCurrencies[currency], fmt.Sprintf("Chapa only supports configured currencies %v, got %s", keysOfBoolMap(chapaSupportedCurrencies), currency))

	callbackURL := sanitizeChapaRedirectURL(webhook2.GetPaymentRedirectEntranceUrlCheckout(createPayContext.Pay, true))
	returnURL := sanitizeChapaRedirectURL(webhook2.GetPaymentRedirectEntranceUrlCheckout(createPayContext.Pay, true))

	payload := map[string]interface{}{
		"amount":       utility.ConvertCentToDollarStr(totalAmount, currency),
		"currency":     currency,
		"email":        createPayContext.Email,
		"first_name":   "UniBee",
		"last_name":    "Customer",
		"tx_ref":       createPayContext.Pay.PaymentId,
		"callback_url": callbackURL,
		"return_url":   returnURL,
		"customization": map[string]interface{}{
			"title":       name,
			"description": description,
		},
	}

	responseJson, err := sendChapaRequest(ctx, apiKey, "POST", "/v1/transaction/initialize", payload)
	log.SaveChannelHttpLog("GatewayNewPayment", payload, responseJson, err, "ChapaNewPayment", nil, gateway)
	if err != nil {
		return nil, err
	}

	utility.Assert(responseJson.Contains("data.checkout_url"), "invalid request, data.checkout_url is nil")
	checkoutURL := responseJson.Get("data.checkout_url").String()
	gatewayPaymentId := responseJson.Get("data.reference").String()
	if len(gatewayPaymentId) == 0 {
		gatewayPaymentId = responseJson.Get("data.chapa_reference").String()
	}
	if len(gatewayPaymentId) == 0 {
		gatewayPaymentId = createPayContext.Pay.PaymentId
	}

	return &gateway_bean.GatewayNewPaymentResp{
		Status:                 consts.PaymentCreated,
		GatewayPaymentId:       gatewayPaymentId,
		GatewayPaymentIntentId: createPayContext.Pay.PaymentId,
		Link:                   checkoutURL,
	}, nil
}

func (c Chapa) GatewayCapture(ctx context.Context, gateway *entity.MerchantGateway, payment *entity.Payment) (res *gateway_bean.GatewayPaymentCaptureResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayCancel(ctx context.Context, gateway *entity.MerchantGateway, payment *entity.Payment) (res *gateway_bean.GatewayPaymentCancelResp, err error) {
	apiKey := c.resolveApiKey(gateway.GatewayKey, gateway.GatewaySecret)
	utility.Assert(len(apiKey) > 0, "Chapa secret key is required")
	utility.Assert(payment != nil && len(payment.PaymentId) > 0, "payment not found")

	_, err = sendChapaRequest(ctx, apiKey, "PUT", fmt.Sprintf("/v1/transaction/cancel/%s", url.PathEscape(payment.PaymentId)), nil)
	if err != nil {
		return nil, err
	}
	return &gateway_bean.GatewayPaymentCancelResp{Status: consts.PaymentCancelled}, nil
}

func (c Chapa) GatewayPaymentList(ctx context.Context, gateway *entity.MerchantGateway, listReq *gateway_bean.GatewayPaymentListReq) (res []*gateway_bean.GatewayPaymentRo, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayPaymentDetail(ctx context.Context, gateway *entity.MerchantGateway, gatewayPaymentId string, payment *entity.Payment) (res *gateway_bean.GatewayPaymentRo, err error) {
	apiKey := c.resolveApiKey(gateway.GatewayKey, gateway.GatewaySecret)
	utility.Assert(len(apiKey) > 0, "Chapa secret key is required")

	txRef := gatewayPaymentId
	if payment != nil && len(payment.PaymentId) > 0 {
		txRef = payment.PaymentId
	}
	utility.Assert(len(txRef) > 0, "tx_ref is empty")

	responseJson, err := sendChapaRequest(ctx, apiKey, "GET", fmt.Sprintf("/v1/transaction/verify/%s", url.PathEscape(txRef)), nil)
	log.SaveChannelHttpLog("GatewayPaymentDetail", map[string]interface{}{"tx_ref": txRef}, responseJson, err, "ChapaPaymentDetail", nil, gateway)
	if err != nil {
		return nil, err
	}

	return parseChapaPayment(responseJson, txRef), nil
}

func (c Chapa) GatewayRefundList(ctx context.Context, gateway *entity.MerchantGateway, gatewayPaymentId string) (res []*gateway_bean.GatewayPaymentRefundResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayRefundDetail(ctx context.Context, gateway *entity.MerchantGateway, gatewayRefundId string, refund *entity.Refund) (res *gateway_bean.GatewayPaymentRefundResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayRefund(ctx context.Context, gateway *entity.MerchantGateway, createPaymentRefundContext *gateway_bean.GatewayNewPaymentRefundReq) (res *gateway_bean.GatewayPaymentRefundResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) GatewayRefundCancel(ctx context.Context, gateway *entity.MerchantGateway, payment *entity.Payment, refund *entity.Refund) (res *gateway_bean.GatewayPaymentRefundResp, err error) {
	return nil, gerror.New("Not Support")
}

func (c Chapa) resolveApiKey(key string, secret string) string {
	if len(strings.TrimSpace(key)) > 0 {
		return strings.TrimSpace(key)
	}
	return strings.TrimSpace(secret)
}

func sanitizeChapaText(value string) string {
	value = chapaDescriptionSanitizer.ReplaceAllString(value, " ")
	value = strings.Join(strings.Fields(value), " ")
	if len(value) == 0 {
		return "UniBee Checkout"
	}
	return value
}

func sanitizeChapaRedirectURL(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "{CHECKOUT_SESSION_ID}", "")
	value = strings.ReplaceAll(value, "session_id=&", "")
	value = strings.ReplaceAll(value, "session_id=", "")
	value = strings.TrimRight(value, "?&")
	return value
}

func keysOfBoolMap(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func parseChapaPayment(item *gjson.Json, txRef string) *gateway_bean.GatewayPaymentRo {
	data := item.GetJson("data")
	statusStr := strings.ToLower(data.Get("status").String())

	status := consts.PaymentCreated
	authorizeStatus := consts.WaitingAuthorized
	switch statusStr {
	case "success", "completed":
		status = consts.PaymentSuccess
		authorizeStatus = consts.Authorized
	case "failed", "failure", "error":
		status = consts.PaymentFailed
	case "cancelled", "canceled", "reversed":
		status = consts.PaymentCancelled
	case "pending":
		status = consts.PaymentCreated
		authorizeStatus = consts.Authorized
	}

	currency := strings.ToUpper(data.Get("currency").String())
	amount := utility.ConvertDollarStrToCent(data.Get("amount").String(), currency)

	gatewayPaymentId := data.Get("chapa_reference").String()
	if len(gatewayPaymentId) == 0 {
		gatewayPaymentId = data.Get("reference").String()
	}
	if len(gatewayPaymentId) == 0 {
		gatewayPaymentId = txRef
	}

	paidTime := parseChapaTime(data.Get("updated_at").String())
	if paidTime == nil {
		paidTime = parseChapaTime(data.Get("created_at").String())
	}

	return &gateway_bean.GatewayPaymentRo{
		GatewayPaymentId: gatewayPaymentId,
		Status:           status,
		AuthorizeStatus:  authorizeStatus,
		AuthorizeReason:  "",
		CancelReason:     "",
		PaymentData:      item.String(),
		Currency:         currency,
		TotalAmount:      amount,
		PaymentAmount:    amount,
		PaidTime:         paidTime,
		Reason:           data.Get("status").String(),
	}
}

func parseChapaTime(value string) *gtime.Time {
	if len(value) == 0 {
		return nil
	}
	if t, err := gtime.StrToTime(value); err == nil {
		return t
	}
	return nil
}

func sendChapaRequest(ctx context.Context, apiKey string, method string, urlPath string, param map[string]interface{}) (res *gjson.Json, err error) {
	utility.Assert(len(strings.TrimSpace(apiKey)) > 0, "api key is nil")

	body := []byte{}
	if param != nil {
		jsonData, marshalErr := gjson.Marshal(param)
		utility.Assert(marshalErr == nil, fmt.Sprintf("json format error %s", marshalErr))
		body = jsonData
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", strings.TrimSpace(apiKey)),
	}

	response, err := utility.SendRequest(chapaAPIBaseURL+urlPath, method, body, headers)
	if err != nil {
		return nil, err
	}

	responseJson, err := gjson.LoadJson(string(response))
	if err != nil {
		return nil, err
	}

	apiStatus := strings.ToLower(responseJson.Get("status").String())
	if len(apiStatus) > 0 && apiStatus != "success" {
		message := responseJson.Get("message").String()
		if len(message) == 0 {
			message = responseJson.String()
		}
		g.Log().Errorf(ctx, "Chapa request failed method:%s path:%s response:%s", method, urlPath, responseJson.String())
		return nil, gerror.New(message)
	}

	return responseJson, nil
}
