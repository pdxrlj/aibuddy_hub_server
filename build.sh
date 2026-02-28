#!/bin/bash

# ==================== 配置区域 ====================
APP_NAME="aibuddy_hub"
IMAGE_NAME="${APP_NAME}:latest"
CONTAINER_NAME="${APP_NAME}_container"

# 端口配置
APP_PORT="9081"              # 应用内部端口
HOST_PORT="9081"             # 本地宿主机映射端口
REMOTE_HOST_PORT="9081"      # 远程宿主机映射端口

# rclone 远程配置
RCLONE_REMOTE="ipaiserver2"          # rclone 配置名称 (默认 dev)

# 远程服务器配置
REMOTE_IP="8.153.82.116"
REMOTE_USER="root"
REMOTE_SSH_PORT="22"         # SSH 端口
REMOTE_DIR="/root/aibuddy_hub_server/${APP_NAME}"
REMOTE_IMAGE_NAME="${APP_NAME}:latest"

# Docker 构建参数
DOCKERFILE="deploy/Dockerfile.dev"
BUILD_CONTEXT="."

# ==================== 颜色输出 ====================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# ==================== SSH 远程连接 ====================
ssh_connect() {
    local command="$1"
    if [ -z "$command" ]; then
        log_info "连接到远程服务器 $REMOTE_USER@$REMOTE_IP..."
        ssh -p "$REMOTE_SSH_PORT" "$REMOTE_USER@$REMOTE_IP"
    else
        log_info "在远程服务器执行: $command"
        ssh -p "$REMOTE_SSH_PORT" "$REMOTE_USER@$REMOTE_IP" "$command"
    fi
}

# ==================== 文件传输 ====================
transfer() {
    local src="$1"
    local dest="$2"
    local remote="${3:-$RCLONE_REMOTE}"
    
    if [ -z "$src" ] || [ -z "$dest" ]; then
        log_error "用法: transfer <src> <dest> [remote]"
        log_info "示例: transfer ./main.go /data/www/app/"
        log_info "      transfer ./main.go /data/www/app/ prod"
        return 1
    fi
    
    log_info "传输文件: $src -> ${remote}:${dest}"
    rclone copy "$src" "${remote}:${dest}" --progress
}

# ==================== 本地 Docker 编译 ====================
docker_build() {
    log_info "开始本地 Docker 构建..."
    
    if [ ! -f "$DOCKERFILE" ]; then
        log_error "Dockerfile 不存在: $DOCKERFILE"
        return 1
    fi
    
    docker build -t "$IMAGE_NAME" -f "$DOCKERFILE" "$BUILD_CONTEXT"
    
    if [ $? -eq 0 ]; then
        log_info "构建成功: $IMAGE_NAME"
        docker images | grep "$APP_NAME"
    else
        log_error "构建失败"
        return 1
    fi
}

# ==================== 远程 Docker 编译 ====================
remote_docker_build() {
    local remote="${1:-$RCLONE_REMOTE}"
    log_info "开始远程 Docker 构建 [rclone: $remote]..."
    
    # 1. 同步代码到远程 (使用 rclone)
    log_info "同步代码到远程服务器..."
    
    # 确保远程目录存在
    rclone mkdir "${remote}:${REMOTE_DIR}" 2>/dev/null
    
    # 使用 rclone sync 同步代码
    rclone sync ./ "${remote}:${REMOTE_DIR}/" \
        --exclude ".git/**" \
        --exclude "logs/**" \
        --exclude "*.exe" \
        --exclude "*.out" \
        --exclude ".idea/**" \
        --exclude ".vscode/**" \
        --progress
    
    if [ $? -ne 0 ]; then
        log_error "代码同步失败"
        return 1
    fi
    
    # 2. 远程构建
    log_info "在远程服务器执行 Docker 构建..."
    ssh_connect "cd $REMOTE_DIR && docker build -t $REMOTE_IMAGE_NAME -f $DOCKERFILE ."
    
    if [ $? -eq 0 ]; then
        log_info "远程构建成功: $REMOTE_IMAGE_NAME"
    else
        log_error "远程构建失败"
        return 1
    fi
}

# ==================== Docker 运行 ====================
docker_run() {
    log_info "启动容器: $CONTAINER_NAME"
    
    # 停止并删除已存在的容器
    docker stop "$CONTAINER_NAME" 2>/dev/null
    docker rm "$CONTAINER_NAME" 2>/dev/null
    
    # 构建 docker run 命令
    local env_opts=""
    
    # 读取 .env 文件
    if [ -f ".env" ]; then
        log_info "加载 .env 环境变量..."
        env_opts="--env-file $(pwd)/.env"
    fi
    
    # 运行新容器
    docker run -d \
        --name "$CONTAINER_NAME" \
        --restart=unless-stopped \
        -p "${HOST_PORT}:${APP_PORT}" \
        -v "$(pwd)/config:/app/config:ro" \
        -v "$(pwd)/logs:/app/logs" \
        $env_opts \
        "$IMAGE_NAME"
    
    if [ $? -eq 0 ]; then
        log_info "容器启动成功"
        docker ps | grep "$CONTAINER_NAME"
    else
        log_error "容器启动失败"
        return 1
    fi
}

# ==================== Docker 重启 ====================
docker_restart() {
    log_info "重启容器: $CONTAINER_NAME"
    docker restart "$CONTAINER_NAME"
    
    if [ $? -eq 0 ]; then
        log_info "容器重启成功"
        docker ps | grep "$CONTAINER_NAME"
    else
        log_error "容器重启失败，尝试重新运行..."
        docker_run
    fi
}

# ==================== Docker 停止 ====================
docker_stop() {
    log_info "停止容器: $CONTAINER_NAME"
    docker stop "$CONTAINER_NAME" 2>/dev/null
    docker rm "$CONTAINER_NAME" 2>/dev/null
    log_info "容器已停止并删除"
}

# ==================== 查看日志 ====================
docker_logs() {
    local lines="${1:-100}"
    docker logs --tail "$lines" -f "$CONTAINER_NAME"
}

# ==================== 远程部署 ====================
remote_deploy() {
    local remote="${1:-$RCLONE_REMOTE}"
    log_info "开始远程部署 [rclone: $remote]..."
    
    # 1. 构建镜像
    remote_docker_build "$remote" || return 1
    
    # 2. 停止旧容器
    ssh_connect "docker stop $CONTAINER_NAME 2>/dev/null; docker rm $CONTAINER_NAME 2>/dev/null"
    
    # 3. 启动新容器 (使用远程 .env 文件)
    ssh_connect "docker run -d --name $CONTAINER_NAME --restart=unless-stopped -p ${REMOTE_HOST_PORT}:${APP_PORT} -v $REMOTE_DIR/config:/app/config:ro -v $REMOTE_DIR/logs:/app/logs --env-file $REMOTE_DIR/.env $REMOTE_IMAGE_NAME"
    
    if [ $? -eq 0 ]; then
        log_info "远程部署成功!"
        ssh_connect "docker ps | grep $CONTAINER_NAME"
    else
        log_error "远程部署失败"
        return 1
    fi
}

# ==================== 远程容器管理 ====================
remote_restart() {
    log_info "远程重启容器: $CONTAINER_NAME"
    ssh_connect "docker restart $CONTAINER_NAME && docker ps | grep $CONTAINER_NAME"
}

remote_stop() {
    log_info "远程停止容器: $CONTAINER_NAME"
    ssh_connect "docker stop $CONTAINER_NAME 2>/dev/null; docker rm $CONTAINER_NAME 2>/dev/null"
}

remote_logs() {
    local lines="${1:-100}"
    ssh_connect "docker logs --tail $lines -f $CONTAINER_NAME"
}

# ==================== 完整部署流程 ====================
full_deploy() {
    log_info "执行完整部署流程..."
    docker_build && docker_run
}

# ==================== 帮助信息 ====================
show_help() {
    echo "
用法: $0 <命令> [参数]

本地操作:
  build         本地 Docker 编译
  run           启动本地 Docker 容器
  restart       重启本地 Docker 容器
  stop          停止并删除本地 Docker 容器
  logs [n]      查看本地容器日志 (默认100行)
  full          完整本地部署 (构建 + 运行)

远程操作:
  remote-build [remote]  远程 Docker 编译 (默认使用 dev 配置)
  deploy [remote]        远程部署 (同步 + 构建 + 运行)
  remote-restart         远程重启容器
  remote-stop            远程停止容器
  remote-logs [n]        远程容器日志

SSH/传输:
  ssh                    SSH 连接到远程服务器
  ssh-cmd <cmd>          在远程服务器执行命令
  transfer <src> <dest> [remote]  传输文件到远程服务器

其他:
  help          显示此帮助信息

rclone 配置:
  默认远程配置名称: $RCLONE_REMOTE
  可通过参数指定其他配置: $0 deploy prod

示例:
  $0 build              # 本地构建镜像
  $0 run                # 启动本地容器
  $0 deploy             # 远程完整部署 (使用默认 dev 配置)
  $0 deploy prod        # 使用 prod 配置远程部署
  $0 remote-build prod  # 使用 prod 配置远程构建
  $0 transfer ./a.txt /tmp/
"
}

# ==================== 主入口 ====================
case "$1" in
    build)
        docker_build
        ;;
    run)
        docker_run
        ;;
    restart)
        docker_restart
        ;;
    stop)
        docker_stop
        ;;
    logs)
        docker_logs "$2"
        ;;
    remote-build)
        remote_docker_build "$2"
        ;;
    deploy)
        remote_deploy "$2"
        ;;
    remote-restart)
        remote_restart
        ;;
    remote-stop)
        remote_stop
        ;;
    remote-logs)
        remote_logs "$2"
        ;;
    ssh)
        ssh_connect
        ;;
    ssh-cmd)
        ssh_connect "$2"
        ;;
    transfer)
        transfer "$2" "$3" "$4"
        ;;
    full)
        full_deploy
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        show_help
        exit 1
        ;;
esac
