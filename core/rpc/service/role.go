package service

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	coreI18n "oa.98ent.com/p9/core/common/i18n"
	"oa.98ent.com/p9/core/rpc/casbinx"
	"oa.98ent.com/p9/core/rpc/ent"
	"oa.98ent.com/p9/core/rpc/ent/role"
	"oa.98ent.com/p9/core/rpc/model"
)

type CreateRoleReq struct {
	RoleCode    string `json:"role_code"`
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
	Status      int16  `json:"status"`
	SortNo      int    `json:"sort_no"`
}

type UpdateRoleReq struct {
	ID          int64   `json:"id"`
	RoleName    *string `json:"role_name"`
	Description *string `json:"description"`
	Status      *int16  `json:"status"`
	SortNo      *int    `json:"sort_no"`
}

func (d *Deps) CreateRole(ctx context.Context, claims *ctxdata.Claims, req CreateRoleReq) (*model.Role, error) {
	req.RoleCode = strings.TrimSpace(req.RoleCode)
	req.RoleName = strings.TrimSpace(req.RoleName)
	if req.RoleCode == "" || req.RoleName == "" {
		return nil, xerr.BadRequest(coreI18n.RoleCodeNameRequired)
	}
	if req.Status == 0 {
		req.Status = model.StatusNormal
	}
	if d.Mode == ModeOn && (claims == nil || claims.OperatorCode == "") {
		return nil, xerr.Unauthorized(coreI18n.Unauthorized)
	}
	// 检查角色编码是否已存在
	oldRole, err := d.GetRoleByCode(ctx, claims, req.RoleCode)
	if err != nil && !ent.IsNotFound(err) {
		return nil, xerr.EntInternalServerError(coreI18n.RoleCreateFailed, err)
	}
	if oldRole != nil {
		return nil, xerr.BadRequest(coreI18n.RoleCodeAlreadyExists)
	}
	// 检查角色名称是否已存在
	oldRole, err = d.GetRoleByName(ctx, claims, req.RoleName)
	if err != nil && !ent.IsNotFound(err) {
		return nil, xerr.EntInternalServerError(coreI18n.RoleCreateFailed, err)
	}
	if oldRole != nil {
		return nil, xerr.BadRequest(coreI18n.RoleNameAlreadyExists)
	}
	// 创建角色
	row, err := d.Client.Role.Create().
		SetRoleCode(req.RoleCode).
		SetRoleName(req.RoleName).
		SetNillableDescription(strPtr(req.Description)).
		SetStatus(req.Status).
		SetIsSystem(false).
		SetSortNo(req.SortNo).
		Save(ctx)
	if err != nil {
		return nil, xerr.EntInternalServerError(coreI18n.RoleCreateFailed, err)
	}
	return roleFromEnt(row), nil
}

func (d *Deps) UpdateRole(ctx context.Context, claims *ctxdata.Claims, req UpdateRoleReq) error {
	r, err := d.mustTenantRole(ctx, claims, req.ID)
	if err != nil {
		return err
	}
	// 系统角色不能禁用
	if r.IsSystem && req.Status != nil && *req.Status == model.StatusDisabled {
		return xerr.Forbidden(coreI18n.RoleCannotDisableSystem)
	}
	// 检查角色名称是否已存在
	if req.RoleName != nil && *req.RoleName != r.RoleName {
		oldRole, err := d.GetRoleByName(ctx, claims, *req.RoleName)
		if err != nil && !ent.IsNotFound(err) {
			return err
		}
		if oldRole != nil {
			return xerr.BadRequest(coreI18n.RoleNameAlreadyExists)
		}
	}

	// 更新角色
	upd := d.Client.Role.UpdateOneID(r.ID)
	if req.RoleName != nil {
		upd.SetRoleName(strings.TrimSpace(*req.RoleName))
	}
	if req.Description != nil {
		upd.SetDescription(*req.Description)
	}
	if req.SortNo != nil {
		upd.SetSortNo(*req.SortNo)
	}
	if req.Status != nil {
		upd.SetStatus(*req.Status)
		if *req.Status == model.StatusDisabled {
			dom := casbinx.Domain(r.OperatorCode)
			if _, err := d.Enforcer.RemoveFilteredPolicy(0, r.RoleCode, dom); err != nil {
				return err
			}
		}
	}
	return upd.Exec(ctx)
}

func (d *Deps) DeleteRoles(ctx context.Context, claims *ctxdata.Claims, ids []int64) error {
	for _, id := range ids {
		r, err := d.mustTenantRole(ctx, claims, id)
		if err != nil {
			return err
		}
		if r.IsSystem {
			return xerr.Forbidden(coreI18n.RoleCannotDeleteSystem)
		}
		n, err := d.Client.Role.Query().Where(role.ID(r.ID)).QueryUsers().Count(ctx)
		if err != nil {
			return err
		}
		if n > 0 {
			return xerr.BadRequest(coreI18n.RoleStillBoundToUsers)
		}
		dom := casbinx.Domain(r.OperatorCode)
		if _, err := d.Enforcer.RemoveFilteredPolicy(0, r.RoleCode, dom); err != nil {
			return err
		}
		if err := d.Client.Role.UpdateOneID(r.ID).ClearMenus().Exec(ctx); err != nil {
			return err
		}
		if err := d.Client.Role.DeleteOneID(r.ID).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

// GetRole 根据角色ID获取角色
func (d *Deps) GetRole(ctx context.Context, claims *ctxdata.Claims, id int64) (*model.Role, error) {
	return d.mustTenantRole(ctx, claims, id)
}

// GetRoleByCode 根据角色编码获取角色
func (d *Deps) GetRoleByCode(ctx context.Context, claims *ctxdata.Claims, code string) (*model.Role, error) {
	roleInfo, err := d.Client.Role.Query().Where(role.RoleCode(code)).First(ctx)
	if err != nil {
		return nil, err
	}

	return roleFromEnt(roleInfo), nil
}

// GetRoleByName 根据角色名称获取角色
func (d *Deps) GetRoleByName(ctx context.Context, claims *ctxdata.Claims, name string) (*model.Role, error) {
	roleInfo, err := d.Client.Role.Query().Where(role.RoleName(name)).First(ctx)
	if err != nil {
		return nil, err
	}

	return roleFromEnt(roleInfo), nil
}

type RoleListReq struct {
	PageReq
	RoleName string
}

func (d *Deps) ListRoles(ctx context.Context, claims *ctxdata.Claims, req RoleListReq) ([]model.Role, int64, error) {
	if claims == nil {
		return nil, 0, xerr.Unauthorized(coreI18n.Unauthorized)
	}
	q := d.Client.Role.Query()
	if s := strings.TrimSpace(req.RoleName); s != "" {
		q.Where(role.Or(
			role.RoleNameContains(s),
			role.RoleCodeContains(s),
		))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	req.normalize(50)
	list, err := q.Order(ent.Asc(role.FieldSortNo), ent.Asc(role.FieldID)).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	return rolesFromEnt(list), int64(total), err
}

func (d *Deps) mustTenantRole(ctx context.Context, claims *ctxdata.Claims, id int64) (*model.Role, error) {
	row, err := d.Client.Role.Query().Where(role.ID(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.NotFound(coreI18n.RoleNotFound)
		}
		return nil, err
	}
	r := roleFromEnt(row)
	return r, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
