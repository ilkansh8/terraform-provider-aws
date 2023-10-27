// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource(name="Lifecycle Policy")
func newDataSourceLifecyclePolicy(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceLifecyclePolicy{}, nil
}

const (
	DSNameLifecyclePolicy = "Lifecycle Policy Data Source"
)

type dataSourceLifecyclePolicy struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceLifecyclePolicy) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_opensearchserverless_lifecycle_policy"
}

func (d *dataSourceLifecyclePolicy) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"created_date": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"id": framework.IDAttribute(),
			"last_modified_date": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
				},
			},
			"policy": schema.StringAttribute{
				Computed: true,
			},
			"policy_version": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.LifecyclePolicyType](),
				},
			},
		},
	}
}

func (d *dataSourceLifecyclePolicy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().OpenSearchServerlessClient(ctx)

	var data dataSourceLifecyclePolicyData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findLifecyclePolicyByNameAndType(ctx, conn, data.Name.ValueString(), data.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, DSNameLifecyclePolicy, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	data.ID = flex.StringToFramework(ctx, out.Name)
	data.Description = flex.StringToFramework(ctx, out.Description)
	data.Name = flex.StringToFramework(ctx, out.Name)
	data.Type = flex.StringValueToFramework(ctx, out.Type)
	data.PolicyVersion = flex.StringToFramework(ctx, out.PolicyVersion)

	createdDate := time.UnixMilli(aws.ToInt64(out.CreatedDate))
	data.CreatedDate = flex.StringValueToFramework(ctx, createdDate.Format(time.RFC3339))

	lastModifiedDate := time.UnixMilli(aws.ToInt64(out.LastModifiedDate))
	data.LastModifiedDate = flex.StringValueToFramework(ctx, lastModifiedDate.Format(time.RFC3339))

	policyBytes, err := out.Policy.MarshalSmithyDocument()

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, DSNameLifecyclePolicy, data.Name.String(), err),
			err.Error(),
		)
	}

	pb := string(policyBytes)
	data.Policy = flex.StringToFramework(ctx, &pb)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceLifecyclePolicyData struct {
	CreatedDate      types.String `tfsdk:"created_date"`
	Description      types.String `tfsdk:"description"`
	ID               types.String `tfsdk:"id"`
	LastModifiedDate types.String `tfsdk:"last_modified_date"`
	Name             types.String `tfsdk:"name"`
	Policy           types.String `tfsdk:"policy"`
	PolicyVersion    types.String `tfsdk:"policy_version"`
	Type             types.String `tfsdk:"type"`
}
