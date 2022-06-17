# kubelet-reloader
a simple reloader to check kubelet status and restart it when running version is different with current version

- kubelet-reloader will watch on /usr/bin/kubelet-new.
- once there is different version `kubelet-new`, the reloader will replace `/usr/bin/kubelet` and restart kubelet.

### Quick Start

```
wget https://github.com/pacoxu/kubeadm-operator/releases/download/v0.1.0/kubelet-reloader-v0.2.0
chmod +x kubelet-reloader-v0.2.0
./kubelet-reloader-v0.2.0
```
