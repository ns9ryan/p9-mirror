package file

import (
	"context"
	"net/http"
	"strings"

	"oa.98ent.com/p9/oss/rpc/internal/logic"
	"oa.98ent.com/p9/oss/rpc/internal/store"
	"oa.98ent.com/p9/oss/rpc/internal/svc"
	pb "oa.98ent.com/p9/oss/rpc/types/oss"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordLogic {
	return &RecordLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RecordLogic) RecordFile(in *pb.FileRecordReq) (*pb.Empty, error) {
	if l.svcCtx.DB == nil {
		return &pb.Empty{}, nil
	}
	if in == nil || strings.TrimSpace(in.GetKey()) == "" {
		return nil, logic.HTTP(http.StatusBadRequest, "key is required")
	}
	l.svcCtx.RememberFile(l.ctx, store.FileRecord{
		ObjectKey: in.GetKey(), Scene: in.GetScene(), Bucket: in.GetBucket(),
		Name: in.GetName(), ContentType: in.GetContentType(), Size: in.GetSize(),
	})
	return &pb.Empty{}, nil
}
