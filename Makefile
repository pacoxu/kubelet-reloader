all: kubeletReloaderContainer

kubeletReloaderContainer:
	docker build -t daocloud.io/daocloud/kubelet-reloader:v0.0.1 .

release: kubeletReloaderContainer
	docker push daocloud.io/daocloud/kubelet-reloader:v0.0.1

.PHONY: all
