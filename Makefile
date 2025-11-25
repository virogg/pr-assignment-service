.PHONY: build lint fmt mocks tests integration-test e2e-test coverage

TARGET ?= pr-assignment-service

build:
	@echo "go build for target ${TARGET}"
	@mkdir -p .bin
	@go build -o ./bin/${TARGET} ./cmd/${TARGET}

lint:
	@echo "Running linter..."
	golangci-lint run --config .golangci.yml ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

mocks:
	@echo "Generating mocks..."
	@mockgen -destination=internal/application/services/pr_service/mocks/mock_user_getter.go -package=mocks ./internal/application/services/pr_service userGetter
	@mockgen -destination=internal/application/services/pr_service/mocks/mock_team_getter.go -package=mocks ./internal/application/services/pr_service teamGetter
	@mockgen -destination=internal/application/services/pr_service/mocks/mock_pr_repo.go -package=mocks ./internal/application/services/pr_service prRepo
	@mockgen -destination=internal/application/services/stats_service/mocks/mock_stats_repo.go -package=mocks ./internal/application/services/stats_service statsRepo
	@mockgen -destination=internal/application/services/team_service/mocks/mock_user_creator.go -package=mocks ./internal/application/services/team_service userCreator
	@mockgen -destination=internal/application/services/team_service/mocks/mock_team_repo.go -package=mocks ./internal/application/services/team_service teamRepo
	@mockgen -destination=internal/application/services/user_service/mocks/mock_user_getter_setter.go -package=mocks ./internal/application/services/user_service userGetterSetter
	@mockgen -destination=internal/application/services/user_service/mocks/mock_team_getter.go -package=mocks ./internal/application/services/user_service teamGetter
	@mockgen -destination=internal/application/services/user_service/mocks/mock_pr_repo.go -package=mocks ./internal/application/services/user_service prRepo
	@mockgen -destination=internal/infrastructure/transport/http/handlers/mocks/mock_pr_service.go -package=mocks ./internal/infrastructure/transport/http/handlers prService
	@mockgen -destination=internal/infrastructure/transport/http/handlers/mocks/mock_stats_service.go -package=mocks ./internal/infrastructure/transport/http/handlers statsService
	@mockgen -destination=internal/infrastructure/transport/http/handlers/mocks/mock_team_service.go -package=mocks ./internal/infrastructure/transport/http/handlers teamService
	@mockgen -destination=internal/infrastructure/transport/http/handlers/mocks/mock_user_service.go -package=mocks ./internal/infrastructure/transport/http/handlers userService

	@echo "Mocks generated successfully"

tests:
	go test ./...

integration-test: mocks
	go test -tags=integration ./...

e2e-test: mocks
	go test -tags=e2e ./...

coverage:
	@go test ./... -coverprofile cover.out.tmp && \
    cat cover.out.tmp \
    	| grep -v "main.go" \
    	| grep -v "mocks/" \
    	| grep -v "pkg/" \
    	| grep -v "repository/" \
    	| grep -v "e2e/" \
    	| grep -v "app.go" \
    	> cover.out && \
    rm cover.out.tmp && \
    go tool cover -func cover.out



