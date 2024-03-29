DOCKER_IMAGE_NAME = docker.servicewall.cn/servicewall/porter:latest
GIT_HASH = $(shell git rev-parse HEAD)
TAR_NAME_PREFIX = porter-$(shell date +'%Y%m')
XZ_THREAD = 3

all:
.PHONY: all makesure_docker_builder docker_push docker_tar_amd64 docker_tar_arm64

build:

push:

makesure_docker_builder:
ifeq ($(shell docker buildx inspect sw_builder 2>/dev/null || echo $$?),1)
	docker buildx create --name sw_builder --platform linux/amd64,linux/arm64 --use
else
	@echo "found docker builder named 'sw_builder'"
	docker buildx use sw_builder
endif

docker_push:
	docker buildx build --platform linux/amd64,linux/arm64 --push \
		-t $(DOCKER_IMAGE_NAME) --build-arg GIT_HASH=$(GIT_HASH) .

docker_tar_amd64: makesure_docker_builder
	set -e; set -o pipefail; \
	docker buildx build --platform linux/amd64 -o type=docker,dest=- \
		-t $(DOCKER_IMAGE_NAME) --build-arg GIT_HASH=$(GIT_HASH) . | xz -T $(XZ_THREAD) > $(TAR_NAME_PREFIX)_amd64.docker.txz
docker_tar_arm64: makesure_docker_builder
	set -e; set -o pipefail; \
	docker buildx build --platform linux/arm64 -o type=docker,dest=- \
		-t $(DOCKER_IMAGE_NAME) --build-arg GIT_HASH=$(GIT_HASH) . | xz -T $(XZ_THREAD) > $(TAR_NAME_PREFIX)_arm64.docker.txz

gen_upgrade_tarball: clean docker_tar_amd64 docker_tar_arm64
	@set -e; \
		export SW_META_VERSION=1.0; \
		export SW_SERVICE_NAME=consoel; \
		export SW_DOCKER_IMAGE_NAME=$(DOCKER_IMAGE_NAME); \
		export SW_META_PLATFORM=amd64; \
		export SW_DOCKER_TAR_FILE=$(TAR_NAME_PREFIX)_$${SW_META_PLATFORM}.docker.txz; \
		export SW_UPGRADE_FILENAME=$${SW_SERVICE_NAME}_upgrade-$(shell date +'%Y%m')_$${SW_META_PLATFORM}.tar; \
		curl -sLf https://res-download.s3.cn-northwest-1.amazonaws.com.cn/antibot/upgrade/scripts/gen_s3_upgrade_tarball.sh | bash; \
		export SW_META_PLATFORM=arm64; \
		export SW_DOCKER_TAR_FILE=$(TAR_NAME_PREFIX)_$${SW_META_PLATFORM}.docker.txz;\
		export SW_UPGRADE_FILENAME=$${SW_SERVICE_NAME}_upgrade-$(shell date +'%Y%m')_$${SW_META_PLATFORM}.tar; \
		curl -sLf https://res-download.s3.cn-northwest-1.amazonaws.com.cn/antibot/upgrade/scripts/gen_s3_upgrade_tarball.sh | bash ;

clean:
	rm -rf *.docker.txz
	rm -rf *.tar

upgrade_tar_amd64: docker_tar_amd64
	@set -e; \
		export SKIP_UPLOAD=true; \
		export SW_META_VERSION=1.0; \
		export SW_SERVICE_NAME=porter; \
		export SW_DOCKER_IMAGE_NAME=$(DOCKER_IMAGE_NAME); \
		export SW_META_PLATFORM=amd64; \
		export SW_DOCKER_TAR_FILE=$(TAR_NAME_PREFIX)_$${SW_META_PLATFORM}.docker.txz; \
		export SW_UPGRADE_FILENAME=$${SW_SERVICE_NAME}_upgrade-$(shell date +'%Y%m')_$${SW_META_PLATFORM}.tar; \
		curl -sLf https://res-download.s3.cn-northwest-1.amazonaws.com.cn/antibot/upgrade/scripts/gen_s3_upgrade_tarball.sh | bash

upgrade_tar_arm64: docker_tar_arm64
	@set -e; \
		export SKIP_UPLOAD=true; \
		export SW_META_VERSION=1.0; \
		export SW_SERVICE_NAME=porter; \
		export SW_DOCKER_IMAGE_NAME=$(DOCKER_IMAGE_NAME); \
		export SW_META_PLATFORM=arm64; \
		export SW_DOCKER_TAR_FILE=$(TAR_NAME_PREFIX)_$${SW_META_PLATFORM}.docker.txz; \
		export SW_UPGRADE_FILENAME=$${SW_SERVICE_NAME}_upgrade-$(shell date +'%Y%m')_$${SW_META_PLATFORM}.tar; \
		curl -sLf https://res-download.s3.cn-northwest-1.amazonaws.com.cn/antibot/upgrade/scripts/gen_s3_upgrade_tarball.sh | bash
