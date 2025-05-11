package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/neptship/calc-yandex-go/internal/config"
	"github.com/neptship/calc-yandex-go/internal/grpc"
	"github.com/neptship/calc-yandex-go/pkg/calculation"
	pb "github.com/neptship/calc-yandex-go/proto"
)

func RunWorker(ctx context.Context, id int, cfg *config.Config, client *grpc.GRPCClient) {
	log.Printf("Worker %d started using gRPC", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return
		default:
			task, err := client.FetchTask(ctx)
			if err != nil {
				log.Printf("Worker %d failed to fetch task: %v", id, err)
				time.Sleep(time.Duration(cfg.AgentPeriodicityMs) * time.Millisecond)
				continue
			}

			if task == nil {
				time.Sleep(time.Duration(cfg.AgentPeriodicityMs) * time.Millisecond)
				continue
			}

			log.Printf("Worker %d processing task ID=%d", id, task.TaskId)

			var arg1, arg2 interface{}

			switch t := task.Arg1.(type) {
			case *pb.TaskResponse_NumberArg1:
				arg1 = t.NumberArg1
			case *pb.TaskResponse_StringArg1:
				arg1 = t.StringArg1
			}

			switch t := task.Arg2.(type) {
			case *pb.TaskResponse_NumberArg2:
				arg2 = t.NumberArg2
			case *pb.TaskResponse_StringArg2:
				arg2 = t.StringArg2
			}

			time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

			result, isError, errorMsg := performOperation(task.Operation, arg1, arg2)

			err = client.SubmitResult(ctx, int(task.TaskId), result, isError, errorMsg)
			if err != nil {
				log.Printf("Worker %d failed to submit result: %v", id, err)
			} else {
				log.Printf("Worker %d submitted result for task ID=%d: %f", id, task.TaskId, result)
			}

			time.Sleep(time.Duration(cfg.AgentPeriodicityMs) * time.Millisecond)
		}
	}
}

func performOperation(operation string, arg1, arg2 interface{}) (float64, bool, string) {
	val1, err1 := convertToFloat64(arg1)
	if err1 != nil {
		return 0, true, "Error converting first argument: " + err1.Error()
	}

	val2, err2 := convertToFloat64(arg2)
	if err2 != nil {
		return 0, true, "Error converting second argument: " + err2.Error()
	}

	result, err := calculation.EvaluateOperation(val1, val2, operation)
	if err != nil {
		return 0, true, "Operation error: " + err.Error()
	}

	return result, false, ""
}

func convertToFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case string:
		var result float64
		_, err := fmt.Sscanf(v, "%f", &result)
		if err != nil {
			return 0, errors.New("could not parse string as float")
		}
		return result, nil
	default:
		return 0, errors.New("unsupported argument type")
	}
}
