package data

import (
	"context"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/utils/v2/dialer"
)

type TenantsRemote struct {
	dialer dialer.IDialer
}

// NewTenantsRemote .
func NewTenantsRemote(
	conf *conf.Bootstrap,
	dm dialer.IDialerManager,
) (ITenantRemote, error) {
	dialer, err := dm.NewServiceDialer("tenants", conf.GetDiscovery().GetTenants())
	if err != nil {
		return nil, err
	}

	return &TenantsRemote{
		dialer: dialer,
	}, nil
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
	ctx context.Context, tenantID, userID int64,
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
