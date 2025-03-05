.PHONY: run-orchestrator run-agent dev

run-all:
	make -j3 run-orchestrator run-agent run-frontend

run-orchestrator:
	go run cmd/orchestrator/main.go

run-agent:
	go run cmd/agent/main.go

run-frontend:
	cd frontend && npm run dev

run-backend:
	make -j2 run-orchestrator run-agent

test:
	go test ./... -v

install:
	go mod tidy
	cd frontend && npm install