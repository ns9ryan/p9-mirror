package vipdefaulticonservicelogic

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
	"oa.98ent.com/p9/platform-base/pkg/resourceurl"
	"oa.98ent.com/p9/platform-base/pkg/rpc/grpcerror"
	"oa.98ent.com/p9/platform-base/rpc/ent/vipdefaulticongroup"
	"oa.98ent.com/p9/platform-base/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/vipdefaulticonpb"
)

type SaveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveLogic {
	return &SaveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Save 按等级组保存默认图标，同一请求内的所有修改一起成功或回滚。
func (l *SaveLogic) Save(
	in *vipdefaulticonpb.SaveVipDefaultIconsRequest,
) (*vipdefaulticonpb.SaveVipDefaultIconsResponse, error) {
	// 修改人来自认证上下文，拒绝分站用户和预览令牌写入总网默认配置。
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	if claims == nil || claims.UserID <= 0 || !claims.IsPlatform || claims.TokenType == ctxdata.TokenTypePreview {
		return nil, status.Error(codes.PermissionDenied, i18nkey.Forbidden)
	}
	// 先校验等级组和图片目录，将已知公开前缀下的 URL 转为 Key，再进入写库事务。
	list, err := validateIconGroups(in, l.svcCtx.Config.ResourceBaseURL)
	if err != nil {
		return nil, err
	}

	// 开启事务，只写入本次明确提交的等级组。
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, enterror.Handle(l.Logger, err)
	}
	// 任一组写入或最终提交失败时，回滚本次请求的全部修改。
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// 本次提交共用修改时间；同一等级组存在则更新，不存在则新增。
	now := time.Now() // 同次保存的所有等级组使用同一修改时间。
	for _, item := range list {
		err := tx.VipDefaultIconGroup.Create().
			SetMinLevel(item.MinLevel).                           // 等级组下限，0 级有效；同时作为唯一键。
			SetMaxLevel(item.MaxLevel).                           // 新增时保存已校验的等级上限，为下限加 9。
			SetIconURL(item.IconUrl).                             // 保存转换后的 Key 或兼容旧 URL，空串表示清空。
			SetUpdatedBy(claims.UserID).                          // 修改人取自认证上下文。
			SetUpdatedAt(now).                                    // 使用本次统一的服务端修改时间。
			OnConflictColumns(vipdefaulticongroup.FieldMinLevel). // 同等级组已存在时更新下面三个字段。
			UpdateIconURL().                                      // 使用上方准备的图片值覆盖旧值，允许清空。
			UpdateUpdatedBy().                                    // 更新为本次登录用户，保留等级范围和首次创建时间。
			UpdateUpdatedAt().                                    // 更新本组修改时间。
			Exec(l.ctx)                                           // 在事务内执行写入，后续 Commit 成功才算保存完成。
		if err != nil {
			return nil, enterror.Handle(l.Logger, err)
		}
	}

	// 提交前任一步失败均回滚，保留既有等级范围和首次创建时间。
	if err := tx.Commit(); err != nil {
		return nil, enterror.Handle(l.Logger, err)
	}
	committed = true
	l.svcCtx.BaseData.VipDefaultIconGroup.Invalidate(l.ctx)

	// 返回空业务响应，是否成功以事务提交结果为准。
	return &vipdefaulticonpb.SaveVipDefaultIconsResponse{}, nil
}

var vipIconPathPattern = regexp.MustCompile(
	`^vipicon/(?:0-9|10-19|20-29|30-39|40-49|50-59|60-69|70-79|80-89|90-99)/[0-9a-f]{32}\.png$`,
)

// validateIconGroups 校验固定范围、重复组和素材归属，按等级升序统一事务写入顺序。
// 返回列表中的 IconUrl 是内部保存值，不直接作为查询响应返回。
func validateIconGroups(
	in *vipdefaulticonpb.SaveVipDefaultIconsRequest, baseURL string,
) ([]*vipdefaulticonpb.VipDefaultIconGroupInfo, error) {
	// 每次允许保存一至十组；未提交的整个等级组保持原值。
	if in == nil || len(in.List) == 0 || len(in.List) > vipDefaultIconGroupCount {
		return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
	}
	seen := make(map[int64]bool, len(in.List))
	list := make([]*vipdefaulticonpb.VipDefaultIconGroupInfo, 0, len(in.List))
	for _, item := range in.List {
		// 检查固定等级范围，并拒绝同一请求重复提交相同等级组。
		if !validIconGroup(item) || seen[item.MinLevel] {
			return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
		}
		// 复用本组图片校验和 Key 转换，转换后的值只用于数据库保存。
		stored, err := iconValue(item, baseURL)
		if err != nil {
			return nil, err
		}
		seen[item.MinLevel] = true
		list = append(list, &vipdefaulticonpb.VipDefaultIconGroupInfo{
			MinLevel: item.MinLevel, // 已校验的等级下限，0 级有效。
			MaxLevel: item.MaxLevel, // 对应等级上限，为下限加 9。
			IconUrl:  stored,        // 转换后的入库值，空串表示清空。
		})
	}
	// 统一按等级升序写入，减少并发请求以不同顺序更新同组记录产生的锁冲突。
	sort.Slice(list, func(left, right int) bool { return list[left].MinLevel < list[right].MinLevel })
	return list, nil
}

// validIconGroup 固定十组范围，0 级有效；空图片由 iconValue 按原有清空语义处理。
func validIconGroup(item *vipdefaulticonpb.VipDefaultIconGroupInfo) bool {
	return item != nil && item.MinLevel >= 0 && item.MinLevel <= 90 &&
		item.MinLevel%vipDefaultIconLevelStep == 0 && item.MaxLevel == item.MinLevel+vipDefaultIconLevelStep-1
}

// iconValue 复用公开地址转换；已知前缀的本组图片存 Key，旧外链和旧目录保持兼容。
func iconValue(item *vipdefaulticonpb.VipDefaultIconGroupInfo, baseURL string) (string, error) {
	// 空字符串表示主动清空该组图片配置，不删除原 OSS 对象。
	value := strings.TrimSpace(item.IconUrl)
	if value == "" {
		return "", nil
	}
	if len(value) > 1024 {
		return "", grpcerror.InvalidArgument(i18nkey.ValidationError)
	}
	// 从完整 URL 中识别 vipicon 目录；不含该目录的旧公开地址按原值兼容。
	key := value
	if !strings.HasPrefix(value, "vipicon/") {
		locator, err := resourceurl.PublicURL(value)
		if err != nil {
			return "", err
		}
		position := strings.LastIndex(locator.Path, "/vipicon/")
		if position < 0 {
			return value, nil
		}
		key = locator.Path[position+1:]
	}
	// 文件必须属于当前等级组，不能把其他等级组的图片保存到本组。
	prefix := fmt.Sprintf("vipicon/%d-%d/", item.MinLevel, item.MaxLevel)
	if !vipIconPathPattern.MatchString(key) || !strings.HasPrefix(key, prefix) {
		return "", grpcerror.InvalidArgument(i18nkey.ValidationError)
	}
	stored, _, err := resourceurl.Value(value, baseURL)
	return stored, err
}
