package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/neptship/calc-yandex-go/internal/orchestrator"
	pb "github.com/neptship/calc-yandex-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type AgentServer struct {
	pb.UnimplementedAgentServiceServer
	service *orchestrator.Service
}

func NewAgentServer(service *orchestrator.Service) *AgentServer {
	return &AgentServer{service: service}
}

func (s *AgentServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.TaskResponse, error) {
	task, err := s.service.GetNextTask()
	if err != nil {
		return nil, err
	}

	response := &pb.TaskResponse{
		TaskId:        int32(task.ID),
		Operation:     task.Operation,
		OperationTime: int32(task.OperationTime),
		ExpressionId:  int32(task.ExpressionID),
	}

	switch v := task.Arg1.(type) {
	case float64:
		response.Arg1 = &pb.TaskResponse_NumberArg1{NumberArg1: v}
	case string:
		response.Arg1 = &pb.TaskResponse_StringArg1{StringArg1: v}
	}

	switch v := task.Arg2.(type) {
	case float64:
		response.Arg2 = &pb.TaskResponse_NumberArg2{NumberArg2: v}
	case string:
		response.Arg2 = &pb.TaskResponse_StringArg2{StringArg2: v}
	}

	return response, nil
}

func (s *AgentServer) SubmitTaskResult(ctx context.Context, req *pb.TaskResultRequest) (*pb.TaskResultResponse, error) {
	var err error
	if req.IsError {
		err = s.service.SetTaskError(int(req.TaskId), req.ErrorMessage)
	} else {
		err = s.service.SetTaskResult(int(req.TaskId), req.Result)
	}

	if err != nil {
		return &pb.TaskResultResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.TaskResultResponse{
		Success: true,
		Message: "Task result saved successfully",
	}, nil
}

func StartGRPCServer(orchService *orchestrator.Service, lis net.Listener) error {
	s := grpc.NewServer()

	pb.RegisterAgentServiceServer(s, NewAgentServer(orchService))

	reflection.Register(s)

	log.Printf("gRPC server is ready to serve on %s", lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}
	return nil
}
