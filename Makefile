.PHONY: run-orchestrator run-agent

run-orchestrator:
 go run cmd/orchestrator/main.go

run-agent:
 go run cmd/agent/main.go

.PHONY: dev
dev:
 make -j2 run-orchestrator run-agent