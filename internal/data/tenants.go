package data

import (
	"context"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/utils/v1/config"
	"gitlab.calendaria.team/services/utils/v1/jwt"
	jwtp "gitlab.calendaria.team/services/utils/v1/jwt"
	"gitlab.calendaria.team/services/utils/v2/dialer"
)

type TenantsRemote struct {
	dialer *dialer.Dialer
	conf   *conf.Bootstrap
	jwt    *jwt.JwtProcessor
}

// NewTenantsRemote .
func NewTenantsRemote(
	conf *conf.Bootstrap,
	c *config.Config,
	jwt *jwtp.JwtProcessor,
) (*TenantsRemote, error) {
	dialer, err := dialer.NewServiceDialer(c, jwt, "tenants", conf.Discovery.Tenants)
	if err != nil {
		return nil, err
	}

	return &TenantsRemote{
		dialer: dialer,
		conf:   conf,
		jwt:    jwt,
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

func (r *TenantsRemote) GetUserTenants(ctx context.Context, actorId int64) ([]*tenants_v1.Tenant, error) {
	client, err := r.getTenantsClient(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := client.ListTenants(ctx, &tenants_v1.ListTenantsRequest{
		ActorId: actorId,
	})
	if err != nil {
		return nil, err
	}

	return reply.GetTenants(), nil
}

func (r *TenantsRemote) GetMemberIdentities(ctx context.Context, tenantId, userId int64) (*tenants_v1.GetMemberIdentitiesReply, error) {
	client, err := r.getMembersClient(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := client.GetMemberIdentities(ctx, &tenants_v1.GetMemberIdentitiesRequest{
		TenantId: tenantId,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	return reply, nil
}
