package grpc

import (
	"context"

	pb "transline.kz/api/proto/customerpb"
)

type Client struct {
	client pb.CustomerServiceClient
}

func New(client pb.CustomerServiceClient) *Client {
	return &Client{client: client}
}

func (c *Client) UpsertCustomer(ctx context.Context, idn string) (*pb.CustomerResponse, error) {
	return c.client.UpsertCustomer(ctx, &pb.UpsertCustomerRequest{Idn: idn})
}
