default: test

DOCKER_FLAGS += --rm
ifeq ($(shell tty > /dev/null && echo 1 || echo 0), 1)
DOCKER_FLAGS += -i
endif

DOCKER := docker
DOCKER_RUN := $(DOCKER) run $(DOCKER_FLAGS)

TERRAFORM_VERSION ?= 1.10.5

EDITORCONFIG_CHECKER_VERSION ?= 3.0.3
EDITORCONFIG_CHECKER := $(DOCKER_RUN) -v=$(CURDIR):/check docker.io/mstruebing/editorconfig-checker:v$(EDITORCONFIG_CHECKER_VERSION)

SHELLCHECK_VERSION ?= 0.10.0
SHELLCHECK := $(DOCKER_RUN) -v=$(CURDIR):/mnt docker.io/koalaman/shellcheck:v$(SHELLCHECK_VERSION)

YAMLLINT_VERSION ?= 0.33.0
YAMLLINT := $(DOCKER_RUN) -v=$(CURDIR):/code docker.io/pipelinecomponents/yamllint:$(YAMLLINT_VERSION) yamllint

GOLANGCI_LINT_VERSION ?= 1.63.4
GOLANGCI_LINT := $(DOCKER_RUN) -v=$(CURDIR):/code -w /code docker.io/golangci/golangci-lint:v$(GOLANGCI_LINT_VERSION) golangci-lint run

lint: lint/terraform lint/editorconfig lint/shellcheck lint/yamllint lint/go

lint/terraform:
	terraform fmt -recursive -check

lint/editorconfig:
	$(EDITORCONFIG_CHECKER)

lint/shellcheck:
	$(SHELLCHECK) $(shell find . -type f -name '*.sh')

lint/yamllint:
	$(YAMLLINT) .

lint/go:
	$(GOLANGCI_LINT) --fix

install:
	go install

cover:
	go tool cover -html=cover.out

test: test/docs test/pkg test/acc

test/docs:
	TFENV_TERRAFORM_VERSION=$(TERRAFORM_VERSION) go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate

test/pkg:
	go test ./pkg/... -v -coverprofile=cover.out $(TESTARGS) -timeout 60m

test/acc:
	TF_ACC=1 TFENV_TERRAFORM_VERSION=$(TERRAFORM_VERSION) go test ./internal/provider/... -v -coverprofile=cover.out $(TESTARGS) -timeout 60m

docs:
	TFENV_TERRAFORM_VERSION=$(TERRAFORM_VERSION) go generate ./...

.PHONY: lint lint/terraform lint/editorconfig lint/shellcheck lint/yamllint lint/go install cover test test/docs test/pkg test/acc docs
