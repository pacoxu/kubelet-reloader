all: kubeletReloaderContainer

kubeletReloaderContainer:
	docker build -t daocloud.io/daocloud/kubelet-reloader:v0.0.2-dev .

release: kubeletReloaderContainer
	docker push daocloud.io/daocloud/kubelet-reloader:v0.0.2-dev

.PHONY: all
