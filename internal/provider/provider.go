// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-hashicups/internal/client"
	"terraform-provider-hashicups/internal/datasources/coffees"
	"terraform-provider-hashicups/internal/resources/order"
)

// Ensure hashicupsProvider satisfies various provider interfaces.
var _ provider.Provider = &hashicupsProvider{}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &hashicupsProvider{
			version: version,
		}
	}
}

// hashicupsProvider is the provider implementation
type hashicupsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// hashicupsProviderModel maps provider schema data to a Go type
type hashicupsProviderModel struct {
	Host       types.String `tfsdk:"host"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	ClientCert types.String `tfsdk:"client_certificate"`
	ClientKey  types.String `tfsdk:"client_private_key"`
	CACert     types.String `tfsdk:"ca_certificate"`
}

// Metadata returns the provider type name.
func (p *hashicupsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hashicups"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *hashicupsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "URI for HashiCups API. May also be provided via HASHICUPS_HOST environment variable.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username for HashiCups API. May also be provided via HASHICUPS_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for HashiCups API. May also be provided via HASHICUPS_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"client_certificate": schema.StringAttribute{
				Description: "PEM-encoded client certificate (mTLS). May also be provided via HASHICUPS_CLIENT_CERT environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"client_private_key": schema.StringAttribute{
				Description: "PEM-encoded private key paired with client_certificate. May also be provided via HASHICUPS_CLIENT_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"ca_certificate": schema.StringAttribute{
				Description: "Optional PEM-encoded CA certificate to trust the HashiCups endpoint. May also be provided via HASHICUPS_CA_CERT environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a HashiCups API client for data sources and resources.
func (p *hashicupsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring HashiCups client")

	// Retrieve provider data from configuration
	var config hashicupsProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// verify all the values are known
	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Hashicup API Host",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown HashiCups API Username",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown HashiCups API Password",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_PASSWORD environment variable.",
		)
	}

	if config.ClientCert.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_certificate"),
			"Unknown HashiCups Client Certificate",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the mTLS client certificate. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_CLIENT_CERT environment variable.",
		)
	}

	if config.ClientKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_private_key"),
			"Unknown HashiCups Client Private Key",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the mTLS private key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_CLIENT_KEY environment variable.",
		)
	}

	if config.CACert.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("ca_certificate"),
			"Unknown HashiCups CA Certificate",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the CA certificate. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_CA_CERT environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but ovverride with Terraform configuration value if set
	host := os.Getenv("HASHICUPS_HOST")
	username := os.Getenv("HASHICUPS_USERNAME")
	password := os.Getenv("HASHICUPS_PASSWORD")
	clientCert := os.Getenv("HASHICUPS_CLIENT_CERT")
	clientKey := os.Getenv("HASHICUPS_CLIENT_KEY")
	caCert := os.Getenv("HASHICUPS_CA_CERT")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.ClientCert.IsNull() {
		clientCert = config.ClientCert.ValueString()
	}

	if !config.ClientKey.IsNull() {
		clientKey = config.ClientKey.ValueString()
	}

	if !config.CACert.IsNull() {
		caCert = config.CACert.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing HashiCups API Host",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API host. "+
				"Set the host value in the configuration or use the HASHICUPS_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing HashiCups API Username",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API username. "+
				"Set the username value in the configuration or use the HASHICUPS_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing HashiCups API Password",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API password. "+
				"Set the password value in the configuration or use the HASHICUPS_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if clientCert == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_certificate"),
			"Missing HashiCups Client Certificate",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the mTLS client certificate. "+
				"Set the client_certificate value in the configuration or use the HASHICUPS_CLIENT_CERT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if clientKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_private_key"),
			"Missing HashiCups Client Private Key",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the mTLS client private key. "+
				"Set the client_private_key value in the configuration or use the HASHICUPS_CLIENT_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "hashicups_host", host)
	ctx = tflog.SetField(ctx, "hashicups_username", username)
	ctx = tflog.SetField(ctx, "hashicups_password", password)
	ctx = tflog.SetField(ctx, "hashicups_client_certificate", clientCert)
	ctx = tflog.SetField(ctx, "hashicups_client_private_key", clientKey)
	ctx = tflog.SetField(ctx, "hashicups_ca_certificate", caCert)

	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "hashicups_password")
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "hashicups_client_certificate", "hashicups_client_private_key", "hashicups_ca_certificate")

	tflog.Debug(ctx, "Creating HashiCups client")

	// Create a new HashiCups client using the configuration values
	apiClient, err := client.New(client.Config{
		Host:       host,
		Username:   username,
		Password:   password,
		ClientCert: clientCert,
		ClientKey:  clientKey,
		CACert:     caCert,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create HashiCups API Client",
			"An unexpected error occurred when creating the HashiCups API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"HashiCups Client Error: "+err.Error(),
		)
		return
	}

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient

	tflog.Info(ctx, "Configured HashiCups client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *hashicupsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		order.NewOrderResource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *hashicupsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		coffees.NewCoffeesDataSource,
	}
}
