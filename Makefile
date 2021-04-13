# import config.env
conf ?= config.env
include $(conf)
export $(shell sed 's/=.*//' $(conf))

# HELP will output the help for each task

.PHONY: help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

# Docker Tasks
# Build the container
build: ## Builds the container with tag "latest"
	docker build -t $(APP_NAME) .

build-nc: ## Build the container with tag "latest" without caching
	docker build --no-cache -t $(APP_NAME) .

run: ## Runs the service on port configured in `config.env`
	docker run -dit -p=$(HOST_PORT):$(CONTAINER_PORT) --name="$(APP_NAME)" $(APP_NAME)

up: build run ## Runs the service on port configured in `config.env` (Alias to run)

stop: ## Stops and remove the running service
	docker stop $(APP_NAME); docker rm $(APP_NAME)