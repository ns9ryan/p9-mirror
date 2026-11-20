package publishdataservicelogic

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	pg "oa.98ent.com/p9/platform-game/rpc/pb/platform_game"

	"oa.98ent.com/p9/operator-game/rpc/ent/game"
	"oa.98ent.com/p9/operator-game/rpc/ent/gamecategory"
	"oa.98ent.com/p9/operator-game/rpc/ent/gamechannel"
	"oa.98ent.com/p9/operator-game/rpc/ent/gameprovider"
	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"
)

type DeleteOperatorDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteOperatorDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOperatorDataLogic {
	return &DeleteOperatorDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除分站的游戏数据
func (l *DeleteOperatorDataLogic) DeleteOperatorData(in *operator_game.DeleteOperatorDataRequest) (*operator_game.DeleteOperatorDataResponse, error) {
	// ==================== 参数校验 ====================

	opCode := strings.TrimSpace(in.OpCode)
	if opCode == "" {
		return &operator_game.DeleteOperatorDataResponse{
			Success: false,
			Message: "分站代码不能为空",
		}, nil
	}

	l.Infof("开始删除分站游戏数据: op_code=%s", opCode)

	// 调用PlatformGameGrpcClient删除总网游戏资源分配
	_, err := l.svcCtx.PlatformGameGrpcClient.GetOperatorGameAllocationServiceClient().DeleteOperatorGame(l.ctx, &pg.DeleteOperatorGameRequest{
		OpCode: opCode,
	})
	if err != nil {
		l.Errorf("删除总网游戏资源分配失败: op_code=%s, error=%v", opCode, err)
		return &operator_game.DeleteOperatorDataResponse{
			Success: false,
			Message: fmt.Sprintf("delete operator game allocation failed: %v", err),
		}, nil
	}

	// ==================== 删除游戏数据 ====================

	// 统计删除数量
	var categoryDeleted, providerDeleted, channelDeleted, gameDeleted int64

	// 删除游戏分类
	if count, err := l.svcCtx.DB.GameCategory.Delete().
		Where(gamecategory.OpCodeEQ(opCode)).
		Exec(l.ctx); err != nil {
		l.Errorf("删除游戏分类失败: op_code=%s, error=%v", opCode, err)
		return &operator_game.DeleteOperatorDataResponse{
			Success: false,
			Message: fmt.Sprintf("delete categories failed: %v", err),
		}, nil
	} else {
		categoryDeleted = int64(count)
		if count > 0 {
			l.Infof("已删除游戏分类: op_code=%s, count=%d", opCode, count)
		}
	}

	// 删除游戏供应商
	if count, err := l.svcCtx.DB.GameProvider.Delete().
		Where(gameprovider.OpCodeEQ(opCode)).
		Exec(l.ctx); err != nil {
		l.Errorf("删除游戏供应商失败: op_code=%s, error=%v", opCode, err)
		return &operator_game.DeleteOperatorDataResponse{
			Success: false,
			Message: fmt.Sprintf("delete providers failed: %v", err),
		}, nil
	} else {
		providerDeleted = int64(count)
		if count > 0 {
			l.Infof("已删除游戏供应商: op_code=%s, count=%d", opCode, count)
		}
	}

	// 删除游戏渠道
	if count, err := l.svcCtx.DB.GameChannel.Delete().
		Where(gamechannel.OpCodeEQ(opCode)).
		Exec(l.ctx); err != nil {
		l.Errorf("删除游戏渠道失败: op_code=%s, error=%v", opCode, err)
		return &operator_game.DeleteOperatorDataResponse{
			Success: false,
			Message: fmt.Sprintf("delete channels failed: %v", err),
		}, nil
	} else {
		channelDeleted = int64(count)
		if count > 0 {
			l.Infof("已删除游戏渠道: op_code=%s, count=%d", opCode, count)
		}
	}

	// 删除游戏
	if count, err := l.svcCtx.DB.Game.Delete().
		Where(game.OpCodeEQ(opCode)).
		Exec(l.ctx); err != nil {
		l.Errorf("删除游戏失败: op_code=%s, error=%v", opCode, err)
		return &operator_game.DeleteOperatorDataResponse{
			Success: false,
			Message: fmt.Sprintf("delete games failed: %v", err),
		}, nil
	} else {
		gameDeleted = int64(count)
		if count > 0 {
			l.Infof("已删除游戏: op_code=%s, count=%d", opCode, count)
		}
	}

	// ==================== 返回删除结果 ====================

	l.Infof("成功删除分站游戏数据: op_code=%s, category=%d, provider=%d, channel=%d, game=%d",
		opCode, categoryDeleted, providerDeleted, channelDeleted, gameDeleted)

	return &operator_game.DeleteOperatorDataResponse{
		Success:         true,
		Message:         "success",
		CategoryDeleted: categoryDeleted,
		ProviderDeleted: providerDeleted,
		ChannelDeleted:  channelDeleted,
		GameDeleted:     gameDeleted,
	}, nil
}
