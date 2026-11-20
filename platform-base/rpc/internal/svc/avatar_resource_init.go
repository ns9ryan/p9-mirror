package svc

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/platform-base/pkg/resourceupload"
)

// mustInitAvatarResources 沿用项目初始化默认数据的方式，仅补齐八张头像和六张头像框。
// 不写死环境相关的 OSS 地址，不覆盖已有图片；首次图片通过修改接口导入。
func (s *ServiceContext) mustInitAvatarResources(ctx context.Context) {
	// 并发初始化可能触发 PostgreSQL 可串行化冲突，包含首次尝试在内最多执行三次。
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = s.initAvatarResources(ctx)
		// 仅重试 40001（事务序列化失败）；成功或其他错误都结束重试。
		var pgError *pgconn.PgError
		if !errors.As(err, &pgError) || pgError.Code != "40001" {
			break
		}
	}
	// 沿用现有启动条件：初始化最终失败时记录致命错误并退出，不继续启动服务。
	logx.Must(err)
}

// initAvatarResources 使用可串行化事务防止多个服务实例同时初始化而增加固定记录数。
func (s *ServiceContext) initAvatarResources(ctx context.Context) error {
	// 真正开启可串行化事务，让数量检查和缺失记录补齐处于同一事务中。
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	// 任一步失败且未成功提交时，回滚本轮新增的记录。
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// 通过事务客户端分别统计头像和头像框，已有记录不覆盖、不替换。
	avatarCount, err := tx.DefaultAvatar.Query().Count(ctx)
	if err != nil {
		return err
	}
	frameCount, err := tx.DefaultAvatarFrame.Query().Count(ctx)
	if err != nil {
		return err
	}
	// 只新增缺少的头像记录；达到或超过八条时保留原记录，不因数量阻止启动。
	// 新记录图片默认为空，后续通过修改接口配置。
	for count := avatarCount; count < resourceupload.AvatarCount; count++ {
		if err := tx.DefaultAvatar.Create().Exec(ctx); err != nil {
			return err
		}
	}
	// 同样补齐头像框，不上传 OSS 文件，也不指定任何站点或会员的默认选择。
	for count := frameCount; count < resourceupload.FrameCount; count++ {
		if err := tx.DefaultAvatarFrame.Create().Exec(ctx); err != nil {
			return err
		}
	}
	// 两类记录全部补齐后一起提交；只有提交成功才关闭延迟回滚。
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}
