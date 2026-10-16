package provider

import (
	"context"

	"oa.98ent.com/p9/oss/api/internal/convert"
	"oa.98ent.com/p9/oss/api/internal/svc"
	"oa.98ent.com/p9/oss/api/internal/types"
	"oa.98ent.com/p9/oss/rpc/ossclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type Logic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Logic {
	return &Logic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *Logic) rpc() (ossclient.Oss, error) {
	return l.svcCtx.Archive()
}

func (l *Logic) List() (*types.ProviderListData, error) {
	cli, err := l.rpc()
	if err != nil {
		return nil, err
	}
	out, err := cli.ListProviders(l.ctx, &ossclient.Empty{})
	if err != nil {
		return nil, err
	}
	return convert.ProviderList(out), nil
}

func (l *Logic) Get(id uint64) (*types.ProviderView, error) {
	cli, err := l.rpc()
	if err != nil {
		return nil, err
	}
	out, err := cli.GetProvider(l.ctx, &ossclient.IDReq{Id: id})
	if err != nil {
		return nil, err
	}
	return convert.Provider(out), nil
}

func (l *Logic) Active() (*types.ProviderView, error) {
	if l.svcCtx.Oss == nil {
		return convert.ProviderFromOSS(l.svcCtx.Settings.OSS, l.svcCtx.ConfigSource, true), nil
	}
	out, err := l.svcCtx.Oss.GetActiveProvider(l.ctx, &ossclient.Empty{})
	if err != nil {
		return nil, err
	}
	if convert.IsProviderSource(out) {
		return convert.Provider(out), nil
	}
	return convert.ProviderFromOSS(l.svcCtx.Settings.OSS, l.svcCtx.ConfigSource, true), nil
}

func (l *Logic) Create(req *types.ProviderView) (*types.ProviderView, error) {
	cli, err := l.rpc()
	if err != nil {
		return nil, err
	}
	out, err := cli.CreateProvider(l.ctx, convert.ProviderReq(req))
	if err != nil {
		return nil, err
	}
	return convert.Provider(out), nil
}

func (l *Logic) Update(req *types.ProviderView) (*types.ProviderView, error) {
	cli, err := l.rpc()
	if err != nil {
		return nil, err
	}
	out, err := cli.UpdateProvider(l.ctx, convert.ProviderReq(req))
	if err != nil {
		return nil, err
	}
	return convert.Provider(out), nil
}

func (l *Logic) Delete(id uint64) error {
	cli, err := l.rpc()
	if err != nil {
		return err
	}
	_, err = cli.DeleteProvider(l.ctx, &ossclient.IDReq{Id: id})
	return err
}

func (l *Logic) Enable(id uint64) (*types.ProviderView, error) {
	cli, err := l.rpc()
	if err != nil {
		return nil, err
	}
	out, err := cli.EnableProvider(l.ctx, &ossclient.IDReq{Id: id})
	if err != nil {
		return nil, err
	}
	return convert.Provider(out), nil
}
