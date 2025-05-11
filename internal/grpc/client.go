package grpc

import (
	"context"
	"fmt"

	pb "github.com/neptship/calc-yandex-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	client pb.AgentServiceClient
	conn   *grpc.ClientConn
}

func NewGRPCClient(address string) (*GRPCClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewAgentServiceClient(conn)
	return &GRPCClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *GRPCClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *GRPCClient) FetchTask(ctx context.Context) (*pb.TaskResponse, error) {
	resp, err := c.client.GetTask(ctx, &pb.GetTaskRequest{})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *GRPCClient) SubmitResult(ctx context.Context, taskID int, result float64, isError bool, errorMsg string) error {
	req := &pb.TaskResultRequest{
		TaskId:       int32(taskID),
		Result:       result,
		IsError:      isError,
		ErrorMessage: errorMsg,
	}

	resp, err := c.client.SubmitTaskResult(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("failed to submit task result: %s", resp.Message)
	}

	return nil
}
