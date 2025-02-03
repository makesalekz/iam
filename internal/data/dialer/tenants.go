package dialer

import (
	"context"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	u_dialer "gitlab.calendaria.team/services/utils/v2/dialer"
)

type ITenantsRemote interface {
	CreateTenants(ctx context.Context, name string) (*tenants_v1.Tenant, error)
	GetUserTenants(ctx context.Context) ([]*tenants_v1.Tenant, error)
	GetMemberIdentities(ctx context.Context, tenantID, userID int64) (*tenants_v1.GetMemberIdentitiesReply, error)
	DeleteUsersTenants(ctx context.Context, usersIDs []int64) error
}

type TenantsRemote struct {
	dialer u_dialer.IDialer
}

// NewTenantsRemote .
func NewTenantsRemote(
	conf *conf.Bootstrap,
	dm u_dialer.IDialerManager,
) (ITenantsRemote, func(), error) {
	dialer, err := dm.NewServiceDialer("tenants", conf.GetDiscovery().GetTenants())
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		dialer.Close()
	}

	return &TenantsRemote{
		dialer: dialer,
	}, cleanup, nil
}

func (r *TenantsRemote) getTenantsClient(ctx context.Context) (tenants_v1.TenantsClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("can't connect to iam: %s", err.Error())
	}

	return tenants_v1.NewTenantsClient(conn), nil
}

func (r *TenantsRemote) getMembersClient(ctx context.Context) (tenants_v1.MembersClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("can't connect to iam: %s", err.Error())
	}

	return tenants_v1.NewMembersClient(conn), nil
}

func (r *TenantsRemote) CreateTenants(ctx context.Context, name string) (*tenants_v1.Tenant, error) {
	client, err := r.getTenantsClient(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := client.CreateTenant(ctx, &tenants_v1.CreateTenantRequest{Name: name})
	if err != nil {
		return nil, err
	}

	return reply.GetTenant(), nil
}

func (r *TenantsRemote) GetUserTenants(ctx context.Context) ([]*tenants_v1.Tenant, error) {
	client, err := r.getTenantsClient(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := client.ListTenants(ctx, &tenants_v1.ListTenantsRequest{})
	if err != nil {
		return nil, err
	}

	return reply.GetTenants(), nil
}

func (r *TenantsRemote) GetMemberIdentities(
	ctx context.Context,
	tenantID, userID int64,
) (*tenants_v1.GetMemberIdentitiesReply, error) {
	client, err := r.getMembersClient(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := client.GetMemberIdentities(
		ctx, &tenants_v1.GetMemberIdentitiesRequest{
			TenantId: tenantID,
			UserId:   userID,
		},
	)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (r *TenantsRemote) DeleteUsersTenants(ctx context.Context, usersIDs []int64) error {
	client, err := r.getTenantsClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.DeleteUsersTenants(ctx, &tenants_v1.DeleteUsersTenantsRequest{UsersIds: usersIDs})
	if err != nil {
		return err
	}

	return nil
}
