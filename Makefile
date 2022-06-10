all: kubeletReloaderContainer

kubeletReloaderContainer:
	docker build -t daocloud.io/daocloud/kubelet-reloader:v0.0.1-dev .

release: kubeletReloaderContainer
	docker push daocloud.io/daocloud/kubelet-reloader:v0.0.1-dev

clean:
	rm -rf pkg
	rm -rf bin
	rm -rf kubelet-reloader

.PHONY: all clean
