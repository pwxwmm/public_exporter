# publice_exporter

`publice_exporter` 是一个基于 Prometheus 的 Exporter，用于采集 **NPU 设备状态** 和 **Docker 运行时信息**，并通过 `/metrics` 接口暴露 Prometheus 指标。  
该项目支持 **Kubernetes 部署**，并计划未来开源，支持更多采集器的扩展。

---

## **📌 功能**
1. **NPU 设备网络接口状态监控**
   - 采集 `hccn_tool -i %d -link_stat -g` 命令输出
   - 统计 **link up / link down 总次数**
   - 计算 **最近两次 link 状态变更时间差**
   
2. **Docker 运行时信息采集**
   - 采集 `docker info | grep 'Default Runtime'`
   - 监控 **默认 Docker 运行时 (runc, kata, etc.)**

3. **标准 Prometheus Exporter**
   - 指标以 `/metrics` 方式暴露
   - 可扩展更多采集器，如 GPU、磁盘、网络等

---

## **📌 安装 & 运行**
### **1️⃣ 从源码运行**
```sh
git clone https://github.com/pwxwmm/publice_exporter.git
cd publice_exporter

go mod tidy
go run cmd/main.go

默认监听 5535 端口，并暴露 /metrics。

2️⃣ 使用 Docker 运行

构建 Docker 镜像

docker buildx build --platform linux/amd64,linux/arm64 -t your-registry/publice-exporter:latest .
docker push your-registry/publice-exporter:latest

运行容器

docker run -d -p 5535:5535 \
  -e DEVICE_ID=0 \
  --name publice_exporter your-registry/publice-exporter:latest

访问 http://localhost:5535/metrics 查看采集的 Prometheus 指标。

📌 Kubernetes 部署

1️⃣ 部署到 K8s

创建 k8s/deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: publice-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: publice-exporter
  template:
    metadata:
      labels:
        app: publice-exporter
    spec:
      containers:
      - name: publice-exporter
        image: your-registry/publice-exporter:latest
        ports:
        - containerPort: 5535
        env:
        - name: DEVICE_ID
          value: "0"

应用 Deployment

kubectl apply -f k8s/deployment.yaml

📌 Prometheus 指标

指标	描述	标签
hccn_link_up_total	NPU 网络接口 UP 次数	device
hccn_link_down_total	NPU 网络接口 DOWN 次数	device
hccn_link_change_duration_seconds	最近两次 Link 状态变更时间差	device
docker_default_runtime	Docker 默认 Runtime（如 runc）	runtime

访问 http://localhost:5535/metrics 获取完整数据。

📌 未来扩展

✅ 多架构支持：arm64 / amd64
✅ 更多采集器：支持 CPU、内存、磁盘、网络等
✅ 更完善的 Kubernetes 部署

🚀 欢迎贡献和 Star ⭐️！

---

**📝 这个 `README.md` 已包含：**
- **项目介绍**
- **功能描述**
- **本地运行**
- **Docker 构建 & 运行**
- **Kubernetes 部署**
- **Prometheus 采集指标**
- **未来扩展计划**
