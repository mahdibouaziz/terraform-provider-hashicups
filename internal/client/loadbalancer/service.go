package loadbalancer

import (
	"context"
	"fmt"

	"terraform-provider-hashicups/internal/client"
)

const basePath = "/loadbalancers"

// Service provides typed load balancer operations.
type Service struct {
	c client.HTTPClient
}

func NewService(c client.HTTPClient) *Service {
	return &Service{c: c}
}

func (s *Service) Create(ctx context.Context, payload Payload) (*LoadBalancer, error) {
	var lb LoadBalancer
	if err := s.c.Post(ctx, basePath, payload, &lb); err != nil {
		return nil, err
	}
	return &lb, nil
}

func (s *Service) Get(ctx context.Context, id string) (*LoadBalancer, error) {
	var lb LoadBalancer
	if err := s.c.Get(ctx, fmt.Sprintf("%s/%s", basePath, id), &lb); err != nil {
		return nil, err
	}
	return &lb, nil
}

func (s *Service) Update(ctx context.Context, id string, payload Payload) (*LoadBalancer, error) {
	if err := s.c.Put(ctx, fmt.Sprintf("%s/%s", basePath, id), payload, nil); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.c.Delete(ctx, fmt.Sprintf("%s/%s", basePath, id))
}
