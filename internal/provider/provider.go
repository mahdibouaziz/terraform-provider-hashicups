// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-hashicups/internal/client"
	"terraform-provider-hashicups/internal/datasources/coffees"
	"terraform-provider-hashicups/internal/datasources/nodes"
	"terraform-provider-hashicups/internal/resources/loadbalancer"
	"terraform-provider-hashicups/internal/resources/order"
	"terraform-provider-hashicups/internal/resources/vm"
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
	Env                 types.String `tfsdk:"env"`
	ClientID            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	ClientCert          types.String `tfsdk:"client_certificate"`
	ClientKey           types.String `tfsdk:"client_private_key"`
	ClientKeyPassphrase types.String `tfsdk:"client_key_passphrase"`
	CACert              types.String `tfsdk:"ca_certificate"`
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
			"env": schema.StringAttribute{
				Description: "Target environment for the API (prod or preprod). May also be provided via HASHICUPS_ENV environment variable. Defaults to prod.",
				Optional:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "Client ID for API authentication. May also be provided via HASHICUPS_CLIENT_ID environment variable.",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "Client secret for API authentication. May also be provided via HASHICUPS_CLIENT_SECRET environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"client_certificate": schema.StringAttribute{
				Description: "Path (relative to the Terraform working directory or absolute) to a PEM-encoded client certificate for mTLS. May also be provided via HASHICUPS_CLIENT_CERT environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"client_private_key": schema.StringAttribute{
				Description: "Path (relative or absolute) to a PEM-encoded private key paired with client_certificate. May also be provided via HASHICUPS_CLIENT_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"client_key_passphrase": schema.StringAttribute{
				Description: "Optional passphrase to decrypt the client private key if it is encrypted. May also be provided via HASHICUPS_CLIENT_KEY_PASSPHRASE environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"ca_certificate": schema.StringAttribute{
				Description: "Optional path (relative or absolute) to a PEM-encoded CA certificate to trust the HashiCups endpoint. May also be provided via HASHICUPS_CA_CERT environment variable.",
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
	if config.Env.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("env"),
			"Unknown HashiCups Environment",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the environment. "+
				"Set env to \"prod\" or \"preprod\", or use the HASHICUPS_ENV environment variable.",
		)
	}
	if config.ClientID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Unknown Client ID",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the client_id. "+
				"Set it in the configuration or via HASHICUPS_CLIENT_ID.",
		)
	}

	if config.ClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Unknown Client Secret",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the client_secret. "+
				"Set it in the configuration or via HASHICUPS_CLIENT_SECRET.",
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

	if config.ClientKeyPassphrase.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_key_passphrase"),
			"Unknown Client Key Passphrase",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the client key passphrase. "+
				"Set it in the configuration or via HASHICUPS_CLIENT_KEY_PASSPHRASE.",
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
	env := os.Getenv("HASHICUPS_ENV")
	clientID := os.Getenv("HASHICUPS_CLIENT_ID")
	clientSecret := os.Getenv("HASHICUPS_CLIENT_SECRET")
	clientCert := os.Getenv("HASHICUPS_CLIENT_CERT")
	clientKey := os.Getenv("HASHICUPS_CLIENT_KEY")
	clientPassphrase := os.Getenv("HASHICUPS_CLIENT_KEY_PASSPHRASE")
	caCert := os.Getenv("HASHICUPS_CA_CERT")

	if !config.Env.IsNull() {
		env = config.Env.ValueString()
	}

	if !config.ClientID.IsNull() {
		clientID = config.ClientID.ValueString()
	}

	if !config.ClientSecret.IsNull() {
		clientSecret = config.ClientSecret.ValueString()
	}

	if !config.ClientCert.IsNull() {
		clientCert = config.ClientCert.ValueString()
	}

	if !config.ClientKey.IsNull() {
		clientKey = config.ClientKey.ValueString()
	}

	if !config.ClientKeyPassphrase.IsNull() {
		clientPassphrase = config.ClientKeyPassphrase.ValueString()
	}

	if !config.CACert.IsNull() {
		caCert = config.CACert.ValueString()
	}

	// Resolve certificate inputs which are expected to be file paths (relative to
	// the Terraform working directory) or raw PEM strings for backward compatibility.
	clientCertContent, err := loadPEMContent(clientCert)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_certificate"),
			"Invalid Client Certificate Path",
			err.Error(),
		)
	}
	clientKeyContent, err := loadPEMContent(clientKey)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_private_key"),
			"Invalid Client Private Key Path",
			err.Error(),
		)
	}
	caCertContent, err := loadPEMContent(caCert)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("ca_certificate"),
			"Invalid CA Certificate Path",
			err.Error(),
		)
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if strings.TrimSpace(env) == "" {
		env = "prod"
	}

	if clientID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing Client ID",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the client_id. "+
				"Set the client_id value in the configuration or use the HASHICUPS_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Missing Client Secret",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the client_secret. "+
				"Set the client_secret value in the configuration or use the HASHICUPS_CLIENT_SECRET environment variable. "+
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

	host, err := resolveHost(env)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("env"),
			"Invalid HashiCups Environment",
			err.Error(),
		)
		return
	}

	ctx = tflog.SetField(ctx, "hashicups_host", host)
	ctx = tflog.SetField(ctx, "hashicups_client_id", clientID)
	ctx = tflog.SetField(ctx, "hashicups_client_secret", clientSecret)
	ctx = tflog.SetField(ctx, "hashicups_client_certificate", clientCert)
	ctx = tflog.SetField(ctx, "hashicups_client_private_key", clientKey)
	ctx = tflog.SetField(ctx, "hashicups_ca_certificate", caCert)

	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "hashicups_client_secret")
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "hashicups_client_certificate", "hashicups_client_private_key", "hashicups_ca_certificate")

	tflog.Debug(ctx, "Creating HashiCups client")

	// Create a new HashiCups client using the configuration values
	apiClient, err := client.New(client.Config{
		Host:                host,
		ClientID:            clientID,
		ClientSecret:        clientSecret,
		ClientCert:          clientCertContent,
		ClientKey:           clientKeyContent,
		ClientKeyPassphrase: clientPassphrase,
		CACert:              caCertContent,
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

// loadPEMContent returns PEM content. If the input contains PEM headers, it is
// returned as-is for backward compatibility; otherwise the string is treated as
// a path (relative to the current working directory if not absolute) and the
// file contents are returned.
func loadPEMContent(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}

	if strings.Contains(value, "-----BEGIN") {
		return value, nil
	}

	path := value
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = filepath.Join(wd, value)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

const apiBase = "https://api-platform-mtls.cib.echonet"

func resolveHost(env string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "prod":
		return apiBase + "/ito-prod-rose-robotics/v1", nil
	case "preprod":
		return apiBase + "/ito-prod-rose-robotics-preprod/v1", nil
	default:
		return "", fmt.Errorf("env must be \"prod\" or \"preprod\"")
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *hashicupsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		order.NewOrderResource,
		vm.NewVMResource,
		loadbalancer.NewLoadBalancerResource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *hashicupsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		coffees.NewCoffeesDataSource,
		nodes.NewNodesDataSource,
	}
}
