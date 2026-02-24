package node

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"terraform-provider-hashicups/internal/client"
)

const nodeBasePath = "/node"

// Service provides typed operations for nodes.
type Service struct {
	c client.HTTPClient
}

func NewService(c client.HTTPClient) *Service {
	return &Service{c: c}
}

// List returns nodes matching the provided filters.
func (s *Service) List(ctx context.Context, f Filters) (*ListResponse, error) {
	query := url.Values{}

	if f.Mock != nil {
		query.Set("mock", strconv.FormatBool(*f.Mock))
	}
	if f.Ecosystem != "" {
		query.Set("ecosystem", f.Ecosystem)
	}
	if f.IsSubstituted != nil {
		query.Set("is_substituted", strconv.FormatBool(*f.IsSubstituted))
	}
	if f.Env != "" {
		query.Set("env", f.Env)
	}
	if f.State != "" {
		query.Set("state", f.State)
	}
	if f.ToscaID != "" {
		query.Set("tosca_id", f.ToscaID)
	}
	if f.ToscaName != "" {
		query.Set("tosca_name", f.ToscaName)
	}
	if len(f.ToscaTypes) > 0 {
		query.Set("tosca_types", strings.Join(f.ToscaTypes, ","))
	}
	if f.Name != "" {
		query.Set("name", f.Name)
	}
	if f.CreatedAt != "" {
		query.Set("created_at", f.CreatedAt)
	}
	if f.UpdatedAt != "" {
		query.Set("updated_at", f.UpdatedAt)
	}
	if f.Page > 0 {
		query.Set("page", strconv.Itoa(f.Page))
	}
	if f.Size > 0 {
		query.Set("size", strconv.Itoa(f.Size))
	}

	path := nodeBasePath
	if len(query) > 0 {
		path = path + "?" + query.Encode()
	}

	var resp ListResponse
	if err := s.c.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}
	return &resp, nil
}
