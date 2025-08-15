
## Author Information

- **Author**: mmwei3
- **Email**: mmwei3@iflytek.com, 1300042631@qq.com
- **Date**: 2025-03-28

### Optimized Dockerfile:


```dockerfile
# Build stage
FROM golang:1.20 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Install dependencies and build the Go binary
RUN go mod tidy && go build -o public_exporter cmd/main.go

# Final image stage
FROM ubuntu:20.04

# Set the working directory inside the container
WORKDIR /app

# Install necessary packages (e.g., curl for health checks)
RUN apt-get update && apt-get install -y curl

# Copy the compiled Go binary from the builder image
COPY --from=builder /app/public_exporter .

# Expose port for the Prometheus metrics endpoint
EXPOSE 5535

# Set the entrypoint to run the exporter
ENTRYPOINT ["/app/public_exporter"]
```

### Build the Docker Image:

You can use the following command to build the Docker image:

```bash
docker build -t public_exporter .
```

### 1. **Deployment Method 1: Using Docker**

You can run the `public_exporter` inside a Docker container. Here's how you can do it.

#### Run the Docker Container:

```bash
docker run -d --name public_exporter -p 5535:5535 public_exporter
docker run -d -p 5535:5535 --name public_exporter -v $(pwd)/config.yaml:/app/config/config.yaml -v /etc/localtime:/etc/localtime -v /opt/scripts:/opt/scripts -v /var/log/exporter:/var/log/exporter  prometheus_public_exporter
```

This command does the following:
- `-d`: Run the container in detached mode.
- `--name public_exporter`: Names the container `public_exporter`.
- `-p 5535:5535`: Exposes the Prometheus metrics endpoint on port 5535.
- `public_exporter`: The name of the image you built.

#### Verify the Running Container:

You can check the status of the container with:

```bash
docker ps
```

To check the logs of the running container:

```bash
docker logs public_exporter
```

#### Optional: Configure Docker to restart automatically

You can configure the container to restart automatically in case of failures or after system reboots. Add the `--restart` flag when running the container:

```bash
docker run -d -p 5535:5535 --name public_exporter -v $(pwd)/config.yaml:/app/config/config.yaml -v /etc/localtime:/etc/localtime -v /opt/scripts:/opt/scripts -v /var/log/exporter:/var/log/exporter  prometheus_public_exporter
```

This will ensure that the container restarts unless it is explicitly stopped.

---

### 2. **Deployment Method 2: Using `systemd` (Service)**

If you want to run `public_exporter` as a service managed by `systemd`, follow these steps:

#### Step 1: Create a `public_exporter` systemd service file

Create the systemd service file at `/etc/systemd/system/public_exporter.service`:

```bash
sudo nano /etc/systemd/system/public_exporter.service
```

Add the following content to the service file:

```ini
[Unit]
Description=Public Exporter Service
After=network.target

[Service]
ExecStart=/app/public_exporter
WorkingDirectory=/app
Restart=always
User=nobody
Group=nogroup
Environment=CONFIG_FILE=/etc/public_exporter/config.yaml

[Install]
WantedBy=multi-user.target
```

Explanation:
- **ExecStart**: Specifies the command to run the `public_exporter` binary.
- **WorkingDirectory**: Specifies where the `public_exporter` binary is located.
- **Restart**: Ensures that the service will restart on failure or when stopped.
- **User and Group**: Runs the service under the `nobody` user for security purposes.
- **ExecStartPre**: Ensures the log directory exists and sets appropriate permissions.

#### Step 2: Reload systemd and enable the service

Once the service file is created, reload `systemd` to pick up the new service definition:

```bash
sudo systemctl daemon-reload
```

Then, enable and start the `public_exporter` service:

```bash
sudo systemctl enable public_exporter
sudo systemctl start public_exporter
```

#### Step 3: Check the Service Status

To verify if the service is running, use:

```bash
sudo systemctl status public_exporter
```

You can view the logs using:

```bash
sudo journalctl -u public_exporter -f
```

#### Step 4: Stop or Restart the Service

If you need to stop or restart the service, use the following commands:

```bash
sudo systemctl stop public_exporter
sudo systemctl restart public_exporter
```

### 3. **Configuring `systemd` for automatic startup**:

By enabling the service with the `systemctl enable` command, the service will automatically start when the machine reboots.


### Ansible-playbook

```yaml
- name: Deploy and setup public_exporter service
  hosts: exporter_hosts
  become: yes
  vars:
    install_path: /define_app/public_exporter

  tasks:
    - name: 创建部署目录
      file:
        path: "{{ install_path }}/{{ item }}"
        state: directory
        mode: '0755'
      loop:
        - config
        - logs
        - scripts
      tags: [prepare, deploy]

    - name: 拷贝主程序 binary
      copy:
        src: ./public_exporter
        dest: "{{ install_path }}/public_exporter"
        mode: '0755'
      tags: [deploy]

    - name: 拷贝配置文件
      copy:
        src: ./config/config.yaml
        dest: "{{ install_path }}/config/config.yaml"
        mode: '0644'
      tags: [deploy]

    - name: 拷贝脚本文件
      copy:
        src: ./scripts/get_optical_link_status.py
        dest: "{{ install_path }}/scripts/get_optical_link_status.py"
        mode: '0755'
      tags: [deploy]

    - name: 写入 systemd unit 文件
      copy:
        dest: /etc/systemd/system/public_exporter.service
        content: |
          [Unit]
          Description=Public Metrics Exporter
          After=network.target

          [Service]
          Type=simple
          ExecStart={{ install_path }}/public_exporter --config.file={{ install_path }}/config/config.yaml
          Restart=on-failure
          RestartSec=5s
          MemoryLimit=2048M

          [Install]
          WantedBy=multi-user.target
        mode: '0644'
      tags: [service]

    - name: 重载 systemd
      command: systemctl daemon-reload
      tags: [service]

    - name: 启动并设置服务开机启动
      systemd:
        name: public_exporter
        enabled: yes
        state: started
      tags: [service]


```


---