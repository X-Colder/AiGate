#!/bin/bash
set -e

# ========== 配置区 ==========
NETWORK="aigate-net"
MYSQL_CONTAINER="aigate-mysql"
REDIS_CONTAINER="aigate-redis"
APP_CONTAINER="aigate-app"
APP_IMAGE="aigate:latest"

MYSQL_IMAGE="swr.cn-north-4.myhuaweicloud.com/ddn-k8s/gcr.io/ml-pipeline/mysql:8.0.26"
REDIS_IMAGE="swr.cn-north-4.myhuaweicloud.com/ddn-k8s/quay.io/opstree/redis:v7.0.5"

MYSQL_ROOT_PASS="root"
MYSQL_USER="aigate"
MYSQL_PASS="aigate_pass"
MYSQL_DB="aigate"
MYSQL_PORT="3306"
REDIS_PORT="6379"
APP_PORT="8081"
JWT_SECRET="aigate-jwt-secret-key-2026"

# ========== 函数 ==========

log()  { echo -e "\033[32m[AiGate]\033[0m $*"; }
warn() { echo -e "\033[33m[AiGate]\033[0m $*"; }
err()  { echo -e "\033[31m[AiGate]\033[0m $*"; exit 1; }

wait_for_container() {
    local name=$1 max=$2 i=0
    log "等待 $name 就绪..."
    while [ $i -lt $max ]; do
        if docker inspect --format='{{.State.Health.Status}}' "$name" 2>/dev/null | grep -q "healthy"; then
            log "$name 已就绪"
            return 0
        fi
        sleep 2
        i=$((i+1))
    done
    err "$name 在 $((max*2)) 秒内未就绪，请检查日志: docker logs $name"
}

do_start() {
    log "====== 开始部署 AiGate ======"

    # 1. 创建网络
    if ! docker network inspect "$NETWORK" >/dev/null 2>&1; then
        log "创建 Docker 网络: $NETWORK"
        docker network create "$NETWORK"
    fi

    # 2. 启动 MySQL
    if docker ps -a --format '{{.Names}}' | grep -q "^${MYSQL_CONTAINER}$"; then
        log "MySQL 容器已存在，启动..."
        docker start "$MYSQL_CONTAINER" 2>/dev/null || true
    else
        log "启动 MySQL..."
        docker run -d \
            --name "$MYSQL_CONTAINER" \
            --network "$NETWORK" \
            -e MYSQL_ROOT_PASSWORD="$MYSQL_ROOT_PASS" \
            -e MYSQL_DATABASE="$MYSQL_DB" \
            -e MYSQL_USER="$MYSQL_USER" \
            -e MYSQL_PASSWORD="$MYSQL_PASS" \
            -p "$MYSQL_PORT:3306" \
            -v aigate_mysql_data:/var/lib/mysql \
            --health-cmd="mysqladmin ping -h localhost" \
            --health-interval=5s \
            --health-timeout=3s \
            --health-retries=10 \
            --restart unless-stopped \
            "$MYSQL_IMAGE"
    fi
    wait_for_container "$MYSQL_CONTAINER" 30

    # 3. 启动 Redis
    if docker ps -a --format '{{.Names}}' | grep -q "^${REDIS_CONTAINER}$"; then
        log "Redis 容器已存在，启动..."
        docker start "$REDIS_CONTAINER" 2>/dev/null || true
    else
        log "启动 Redis..."
        docker run -d \
            --name "$REDIS_CONTAINER" \
            --network "$NETWORK" \
            -p "$REDIS_PORT:6379" \
            --health-cmd="redis-cli ping" \
            --health-interval=5s \
            --health-timeout=3s \
            --health-retries=5 \
            --restart unless-stopped \
            "$REDIS_IMAGE"
    fi
    wait_for_container "$REDIS_CONTAINER" 15

    # 4. 构建 AiGate 镜像
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
    log "构建 AiGate 镜像..."
    docker build -t "$APP_IMAGE" "$SCRIPT_DIR"

    # 5. 停止旧的 AiGate 容器
    if docker ps -a --format '{{.Names}}' | grep -q "^${APP_CONTAINER}$"; then
        log "移除旧 AiGate 容器..."
        docker rm -f "$APP_CONTAINER" 2>/dev/null || true
    fi

    # 6. 启动 AiGate
    log "启动 AiGate..."
    docker run -d \
        --name "$APP_CONTAINER" \
        --network "$NETWORK" \
        -p "$APP_PORT:8081" \
        -e AIGATE_SERVER_PORT="8081" \
        -e AIGATE_SERVER_MODE="release" \
        -e AIGATE_DB_DRIVER="mysql" \
        -e AIGATE_DB_HOST="$MYSQL_CONTAINER" \
        -e AIGATE_DB_PORT="3306" \
        -e AIGATE_DB_USERNAME="$MYSQL_USER" \
        -e AIGATE_DB_PASSWORD="$MYSQL_PASS" \
        -e AIGATE_DB_DATABASE="$MYSQL_DB" \
        -e AIGATE_REDIS_ADDR="${REDIS_CONTAINER}:6379" \
        -e AIGATE_JWT_SECRET="$JWT_SECRET" \
        -e AIGATE_LOG_LEVEL="info" \
        --restart unless-stopped \
        "$APP_IMAGE"

    log "====== 部署完成 ======"
    log ""
    log "  访问地址:  http://localhost:${APP_PORT}"
    log "  默认账号:  admin / admin123"
    log ""
    log "  查看日志:  docker logs -f $APP_CONTAINER"
    log "  停止服务:  $0 stop"
    log "  清理所有:  $0 clean"
}

do_stop() {
    log "停止所有容器..."
    docker stop "$APP_CONTAINER" "$REDIS_CONTAINER" "$MYSQL_CONTAINER" 2>/dev/null || true
    log "已停止"
}

do_update() {
    log "====== 增量更新 AiGate ======"

    # 1. 拉取最新代码
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
    if [ -d "$SCRIPT_DIR/.git" ]; then
        log "拉取最新代码..."
        git -C "$SCRIPT_DIR" pull --ff-only || warn "代码拉取失败，使用本地代码继续构建"
    fi

    # 2. 重新构建镜像
    log "构建新版本镜像..."
    docker build -t "$APP_IMAGE" "$SCRIPT_DIR"

    # 3. 仅重启 App 容器（保留 MySQL/Redis）
    if docker ps -a --format '{{.Names}}' | grep -q "^${APP_CONTAINER}$"; then
        log "停止旧版本..."
        docker rm -f "$APP_CONTAINER" 2>/dev/null || true
    fi

    # 4. 检查 MySQL/Redis 是否运行中
    if ! docker ps --format '{{.Names}}' | grep -q "^${MYSQL_CONTAINER}$"; then
        log "MySQL 未运行，启动..."
        docker start "$MYSQL_CONTAINER" 2>/dev/null || err "MySQL 容器不存在，请先执行: $0 start"
        wait_for_container "$MYSQL_CONTAINER" 30
    fi
    if ! docker ps --format '{{.Names}}' | grep -q "^${REDIS_CONTAINER}$"; then
        log "Redis 未运行，启动..."
        docker start "$REDIS_CONTAINER" 2>/dev/null || err "Redis 容器不存在，请先执行: $0 start"
        wait_for_container "$REDIS_CONTAINER" 15
    fi

    # 5. 启动新版本
    log "启动新版本..."
    docker run -d \
        --name "$APP_CONTAINER" \
        --network "$NETWORK" \
        -p "$APP_PORT:8081" \
        -e AIGATE_SERVER_PORT="8081" \
        -e AIGATE_SERVER_MODE="release" \
        -e AIGATE_DB_DRIVER="mysql" \
        -e AIGATE_DB_HOST="$MYSQL_CONTAINER" \
        -e AIGATE_DB_PORT="3306" \
        -e AIGATE_DB_USERNAME="$MYSQL_USER" \
        -e AIGATE_DB_PASSWORD="$MYSQL_PASS" \
        -e AIGATE_DB_DATABASE="$MYSQL_DB" \
        -e AIGATE_REDIS_ADDR="${REDIS_CONTAINER}:6379" \
        -e AIGATE_JWT_SECRET="$JWT_SECRET" \
        -e AIGATE_LOG_LEVEL="info" \
        --restart unless-stopped \
        "$APP_IMAGE"

    log "====== 更新完成 ======"
    log ""
    log "  访问地址:  http://localhost:${APP_PORT}"
    log "  数据库已保留，AutoMigrate 自动处理表结构变更"
    log "  查看日志:  docker logs -f $APP_CONTAINER"
    log ""
}

do_clean() {
    warn "将删除所有容器、网络和数据卷，此操作不可逆！"
    read -p "确认删除？[y/N] " confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        log "已取消"
        return
    fi
    log "清理容器..."
    docker rm -f "$APP_CONTAINER" "$REDIS_CONTAINER" "$MYSQL_CONTAINER" 2>/dev/null || true
    log "清理网络..."
    docker network rm "$NETWORK" 2>/dev/null || true
    log "清理数据卷..."
    docker volume rm aigate_mysql_data 2>/dev/null || true
    log "清理镜像..."
    docker rmi "$APP_IMAGE" 2>/dev/null || true
    log "清理完成"
}

do_status() {
    echo ""
    echo "容器状态:"
    echo "-------------------------------------------"
    docker ps -a --filter "name=aigate-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null
    echo ""
}

do_logs() {
    docker logs -f "$APP_CONTAINER"
}

# ========== 入口 ==========
case "${1:-start}" in
    start)   do_start  ;;
    stop)    do_stop   ;;
    update)  do_update ;;
    clean)   do_clean  ;;
    status)  do_status ;;
    logs)    do_logs   ;;
    restart)
        do_stop
        sleep 2
        do_start
        ;;
    *)
        echo "用法: $0 {start|stop|update|restart|status|logs|clean}"
        echo ""
        echo "  start    首次部署（创建 MySQL/Redis/AiGate）"
        echo "  update   增量更新（拉取代码 → 重建镜像 → 仅重启 App，保留数据）"
        echo "  stop     停止所有容器"
        echo "  restart  停止后重新完整部署"
        echo "  status   查看容器状态"
        echo "  logs     查看 AiGate 日志"
        echo "  clean    删除所有容器、网络和数据卷（不可逆）"
        exit 1
        ;;
esac
