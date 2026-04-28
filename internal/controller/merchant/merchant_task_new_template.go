package merchant

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"unibee/api/bean"
	dao "unibee/internal/dao/default"
	_interface "unibee/internal/interface/context"
	entity "unibee/internal/model/entity/default"
	"unibee/utility"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"unibee/api/merchant/task"
)

func (c *ControllerTask) NewTemplate(ctx context.Context, req *task.NewTemplateReq) (res *task.NewTemplateRes, err error) {
	utility.Assert(_interface.Context().Get(ctx).MerchantMember != nil, "No Permission")
	utility.Assert(len(req.Task) > 0, "Invalid Task")
	now := gtime.Now()
	one := &entity.MerchantBatchExportTemplate{
		MerchantId:    _interface.GetMerchantId(ctx),
		MemberId:      _interface.Context().Get(ctx).MerchantMember.Id,
		Name:          req.Name,
		Task:          req.Task,
		Format:        req.Format,
		Payload:       utility.MarshalToJsonString(req.Payload),
		ExportColumns: utility.MarshalToJsonString(req.ExportColumns),
		IsDeleted:     0,
		GmtCreate:     now,
		GmtModify:     now,
		CreateTime:    now.Timestamp(),
	}
	insertData := g.Map{
		dao.MerchantBatchExportTemplate.Columns().MerchantId:    one.MerchantId,
		dao.MerchantBatchExportTemplate.Columns().MemberId:      one.MemberId,
		dao.MerchantBatchExportTemplate.Columns().Name:          one.Name,
		dao.MerchantBatchExportTemplate.Columns().Task:          one.Task,
		dao.MerchantBatchExportTemplate.Columns().Format:        one.Format,
		dao.MerchantBatchExportTemplate.Columns().Payload:       one.Payload,
		dao.MerchantBatchExportTemplate.Columns().ExportColumns: one.ExportColumns,
		dao.MerchantBatchExportTemplate.Columns().IsDeleted:     one.IsDeleted,
		dao.MerchantBatchExportTemplate.Columns().GmtCreate:     one.GmtCreate,
		dao.MerchantBatchExportTemplate.Columns().GmtModify:     one.GmtModify,
		dao.MerchantBatchExportTemplate.Columns().CreateTime:    one.CreateTime,
	}
	dbType := dao.MerchantBatchExportTemplate.DB().GetConfig().Type
	if dbType == "pgsql" || dbType == "postgres" {
		id, err := dao.MerchantBatchExportTemplate.DB().GetValue(ctx, `
INSERT INTO merchant_batch_export_template (
  merchant_id, member_id, name, task, format, payload, export_columns,
  is_deleted, gmt_create, gmt_modify, create_time
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
			one.MerchantId,
			one.MemberId,
			one.Name,
			one.Task,
			one.Format,
			one.Payload,
			one.ExportColumns,
			one.IsDeleted,
			one.GmtCreate,
			one.GmtModify,
			one.CreateTime,
		)
		if err != nil {
			g.Log().Errorf(ctx, "New MerchantBatchExportTemplate Insert err:%s", err.Error())
			return nil, gerror.NewCode(gcode.New(500, "server error", nil))
		}
		one.Id = id.Uint64()
		return &task.NewTemplateRes{Template: bean.SimplifyMerchantBatchExportTemplate(one)}, nil
	}

	id, err := dao.MerchantBatchExportTemplate.Ctx(ctx).Data(insertData).InsertAndGetId()
	if err != nil {
		g.Log().Errorf(ctx, "New MerchantBatchExportTemplate Insert err:%s", err.Error())
		return nil, gerror.NewCode(gcode.New(500, "server error", nil))
	}
	one.Id = uint64(id)

	return &task.NewTemplateRes{Template: bean.SimplifyMerchantBatchExportTemplate(one)}, nil
}
