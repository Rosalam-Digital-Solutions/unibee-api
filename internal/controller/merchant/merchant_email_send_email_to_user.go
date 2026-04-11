package merchant

import (
	"context"
	"fmt"
	"strings"
	"time"
	_interface "unibee/internal/interface/context"
	email2 "unibee/internal/logic/email"
	"unibee/internal/logic/invoice/handler"
	"unibee/internal/query"
	"unibee/utility"

	"unibee/api/merchant/email"
)

func (c *ControllerEmail) SendEmailToUser(ctx context.Context, req *email.SendEmailToUserReq) (res *email.SendEmailToUserRes, err error) {
	user := query.GetUserAccountByEmail(ctx, _interface.GetMerchantId(ctx), req.Email)
	mailTo := strings.ToLower(req.Email)
	language := "en"
	if user != nil && len(user.Language) > 0 {
		language = user.Language
	}
	gatewayName, gatewayData := email2.GetDefaultMerchantEmailConfigWithClusterCloud(ctx, _interface.GetMerchantId(ctx))
	if len(gatewayData) == 0 {
		utility.Assert(false, "Default Email Gateway Need Setup")
	}
	var pdfFileName string
	var attachName string
	if len(req.AttachInvoiceId) == 0 && req.Variables != nil && req.Variables["AttachInvoiceId"] != nil {
		req.AttachInvoiceId = fmt.Sprintf("%s", req.Variables["AttachInvoiceId"])
	}
	if len(req.AttachInvoiceId) > 0 {
		utility.Assert(user != nil, "User not found for attached invoice")
		one := query.GetInvoiceByInvoiceId(ctx, req.AttachInvoiceId)
		utility.Assert(one != nil, "invoice not found")
		utility.Assert(one.UserId > 0 && one.UserId == user.Id, "invoice userId not match")
		pdfFileName = handler.GenerateInvoicePdf(ctx, one)
		attachName = fmt.Sprintf("invoice_%s", time.Now().Format("20060102"))
	}
	err = email2.Send(ctx, &email2.SendEmailReq{
		MerchantId:        _interface.GetMerchantId(ctx),
		MailTo:            mailTo,
		Subject:           req.Subject,
		Content:           req.Content,
		LocalFilePath:     pdfFileName,
		AttachName:        attachName + ".pdf",
		GatewayName:       gatewayName,
		GatewayData:       gatewayData,
		VariableMap:       req.Variables,
		Language:          language,
		GatewayTemplateId: req.GatewayTemplateId,
	})
	if err != nil {
		return nil, err
	}
	return &email.SendEmailToUserRes{}, nil
}
