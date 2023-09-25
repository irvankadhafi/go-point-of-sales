SHELL:=/bin/bash

ifdef test_run
	TEST_ARGS := -run $(test_run)
endif

test_command=richgo test ./... $(TEST_ARGS) -v --cover
migrate_up=go run main.go migrate --direction=up --step=0
migrate_down=go run main.go migrate --direction=down --step=0
run_command=go run main.go server
changelog_args=-o CHANGELOG.md -tag-filter-pattern '^v'

auth/mock/mock_user_authenticator.go:
	mockgen -destination=auth/mock/mock_user_authenticator.go -package=mock github.com/irvankadhafi/go-point-of-sales/auth UserAuthenticator

internal/model/mock/mock_user_repository.go:
	mockgen -destination=internal/model/mock/mock_user_repository.go -package=mock github.com/irvankadhafi/go-point-of-sales/internal/model UserRepository

internal/model/mock/mock_audit_log_repository.go:
	mockgen -destination=internal/model/mock/mock_audit_log_repository.go -package=mock github.com/irvankadhafi/go-point-of-sales/internal/model AuditRepository

internal/model/mock/mock_auth_usecase.go:
	mockgen -destination=internal/model/mock/mock_auth_usecase.go -package=mock github.com/irvankadhafi/go-point-of-sales/internal/model AuthUsecase

internal/model/mock/mock_user_usecase.go:
	mockgen -destination=internal/model/mock/mock_user_usecase.go -package=mock github.com/irvankadhafi/go-point-of-sales/internal/model UserUsecase

internal/model/mock/mock_session_repository.go:
	mockgen -destination=internal/model/mock/mock_session_repository.go -package=mock github.com/irvankadhafi/go-point-of-sales/internal/model SessionRepository

internal/model/mock/mock_rbac_repository.go:
	mockgen -destination=internal/model/mock/mock_rbac_repository.go -package=mock github.com/irvankadhafi/go-point-of-sales/internal/model RBACRepository

internal/model/mock/mock_app_client_usecase.go:
	mockgen -destination=internal/model/mock/mock_app_client_usecase.go -package=mock github.com/irvankadhafi/go-point-of-sales/internal/model AppClientUsecase

internal/model/mock/mock_app_client_repository.go:
	mockgen -destination=internal/model/mock/mock_app_client_repository.go -package=mock github.com/irvankadhafi/go-point-of-sales/internal/model AppClientRepository

mockgen: internal/model/mock/mock_user_repository.go \
	internal/model/mock/mock_audit_log_repository.go \
	internal/model/mock/mock_user_usecase.go \
	internal/model/mock/mock_session_repository.go \
	internal/model/mock/mock_auth_usecase.go \
	internal/model/mock/mock_rbac_repository.go \
	internal/model/mock/mock_app_client_usecase.go \
	internal/model/mock/mock_app_client_repository.go

run: check-modd-exists
	@modd -f ./.modd/server.modd.conf

run-worker: check-modd-exists
	@modd -f ./.modd/worker.modd.conf

run-dlq-worker: check-modd-exists
	@modd -f ./.modd/dlq-worker.modd.conf

check-cognitive-complexity:
	find . -type f -name '*.go' -not -name "*.pb.go" -not -name "mock*.go" -not -name "generated.go" -not -name "federation.go" \
      -exec gocognit -over 15 {} +

lint: check-cognitive-complexity
	golangci-lint run --print-issued-lines=false --exclude-use-default=false --enable=revive --enable=goimports  --enable=unconvert --enable=unparam --concurrency=2

test-only: check-gotest mockgen
	SVC_DISABLE_CACHING=true $(test_command)

test: lint test-only

check-modd-exists:
	@modd --version > /dev/null

check-gotest:
ifeq (, $(shell which richgo))
	$(warning "richgo is not installed, falling back to plain go test")
	$(eval TEST_BIN=go test)
else
	$(eval TEST_BIN=richgo test)
endif

ifdef test_run
	$(eval TEST_ARGS := -run $(test_run))
endif
	$(eval test_command=$(TEST_BIN) ./... $(TEST_ARGS) -v --cover)

migrate:
	@if [ "$(DIRECTION)" = "" ] || [ "$(STEP)" = "" ]; then\
    	$(migrate_up);\
	else\
		go run main.go migrate --direction=$(DIRECTION) --step=$(STEP);\
    fi

clean:
	rm -v internal/model/mock/mock_*.go

changelog:
ifdef version
	$(eval changelog_args=--next-tag $(version) $(changelog_args))
endif
	git-chglog $(changelog_args)


.PHONY: run proto test clean mockgen check-modd-exists
