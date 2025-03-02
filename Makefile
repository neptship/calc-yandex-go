.PHONY: run-orchestrator run-agent dev

run-orchestrator:
	go run cmd/orchestrator/main.go

run-agent:
	go run cmd/agent/main.go

dev:
	make -j2 run-orchestrator run-agent