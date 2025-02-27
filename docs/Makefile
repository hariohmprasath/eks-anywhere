IMAGE_REGISTRY ?= public.ecr.aws/eks-distro-build-tooling
CONTAINER_IMAGE   = $(IMAGE_REGISTRY)/eks-a-hugo
CONTAINER_RUN     = docker run --rm --interactive --tty --volume $(CURDIR)/../:/src
DOCS_ARCHIVE      = public.zip

.PHONY: help build release submodule zip deploy serve server container-build container-serve container-server clean

help: ## Show this help. Requires locally installed npm and hugo.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: submodule ## Generate the static site into the /public folder
	npm install
	hugo --environment production --cleanDestinationDir

release: build

submodule: ## initialize the docsy theme submodule
	git submodule update --init --recursive

submodule-reset: ## reset submodules to tracked commit
	git submodule foreach --recursive git reset --hard

zip: clean build ## Create zip file of assets to upload
	# move into public folder so zip file doesn't have a public folder
	# put the zip file up one directory in docs to be used by deploy target
	cd public && zip -r --quiet ../$(DOCS_ARCHIVE) ./

deploy: zip ## Deploy docs to Amplify
	bash ./deploy-docs.sh public.zip

serve: submodule ## Boot the development server.
	hugo server --buildFuture --baseUrl http://127.0.0.1

server: serve

container-build: submodule ## Build a container image for the preview of the website
	docker build -t $(CONTAINER_IMAGE) .

container-serve: ## Boot the development server using container. Run `make container-build` before this.
	$(CONTAINER_RUN) --mount type=tmpfs,destination=/tmp,tmpfs-mode=01777 \
		--workdir /src/docs \
		-p 1313:1313 $(CONTAINER_IMAGE) server \
		--buildFuture \
		--buildDrafts \
		--bind 0.0.0.0 \
		--destination /tmp/hugo \
		--cleanDestinationDir \
		--baseUrl http://0.0.0.0

container-server: container-serve

clean: ## Delete generated content. (requires sudo if using container-serve)
	rm -rf public resources node_modules public.zip
