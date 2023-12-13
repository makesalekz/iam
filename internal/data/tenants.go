package data

import (
	"context"

	"gitlab.calendaria.team/services/iam/internal/conf"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/utils/v1/dialer"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

type TenantsRemote struct {
	dialer *dialer.Dialer
	conf   *conf.Bootstrap
	jwt    *jwt.JwtProcessor
}

// NewTenantsRemote .
func NewTenantsRemote(d *dialer.Dialer, conf *conf.Bootstrap, jwt *jwt.JwtProcessor) (*TenantsRemote, error) {
	return &TenantsRemote{
		dialer: d,
		conf:   conf,
		jwt:    jwt,
	}, nil
}

func (r *TenantsRemote) GetTenantsClient(ctx context.Context, claims *jwt.TenantClaims) (tenants_v1.TenantsClient, error) {
	return dialer.NewDialerBuilder(r.dialer, tenants_v1.NewTenantsClient).
		SetEndpoint(r.conf.Discovery.Tenants).
		SetTimeout(r.conf.Discovery.TenantsTimeout.AsDuration()).
		Conn(ctx, claims)
}

func (r *TenantsRemote) GetMembersClient(ctx context.Context, claims *jwt.TenantClaims) (tenants_v1.MembersClient, error) {
	return dialer.NewDialerBuilder(r.dialer, tenants_v1.NewMembersClient).
		SetEndpoint(r.conf.Discovery.Tenants).
		SetTimeout(r.conf.Discovery.TenantsTimeout.AsDuration()).
		Conn(ctx, claims)
}

func (r *TenantsRemote) GetUserTenants(ctx context.Context, claims *jwt.TenantClaims) ([]*tenants_v1.Tenant, error) {
	client, err := r.GetTenantsClient(ctx, claims)
	if err != nil {
		return nil, tenants_v1.ErrorGrpcConnection("tenants: %s", err.Error())
	}

	reply, err := client.ListTenants(ctx, &tenants_v1.ListTenantsRequest{})
	if err != nil {
		return nil, err
	}

	return reply.GetTenants(), nil
}

func (r *TenantsRemote) GetMemberIdentities(ctx context.Context, claims *jwt.TenantClaims, userId int64) (*tenants_v1.GetMemberIdentitiesReply, error) {
	client, err := r.GetMembersClient(ctx, claims)
	if err != nil {
		return nil, tenants_v1.ErrorGrpcConnection("tenants: %s", err.Error())
	}

	reply, err := client.GetMemberIdentities(ctx, &tenants_v1.GetMemberIdentitiesRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	return reply, nil
}
