#!/bin/bash
set -euo pipefail


RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}INFO${NC} $1"; }
error() { echo -e "${RED}ERROR${NC} $1"; }
step() { echo -e "\n${YELLOW}=== $1 ===${NC}"; }

BIN_DIR="bin"
K8S_VERSION="v1.34.1"
ETCD_VERSION="v3.6.5"

step "Подготовка"
mkdir -p "$BIN_DIR" || error "Ошибка создания директории $BIN_DIR"
log "Создана директория: $BIN_DIR"

download_and_install() {
    local url="$1"
    local bin_name="$2"
    local target_dir="$3"
    
    step "Скачивание $bin_name"
    curl -L -o "$bin_name.tar.gz" "$url" || error "Ошибка скачивания $bin_name"
    tar -xzf "$bin_name.tar.gz"
    find . -name "$bin_name" -exec mv {} "$target_dir"/ \;
    rm -f "$bin_name.tar.gz"
    log "$bin_name установлен в $target_dir/$bin_name"
}

download_and_install \
    "https://github.com/etcd-io/etcd/releases/download/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz" \
    "etcd" \
    "$BIN_DIR"

download_and_install \
    "https://dl.k8s.io/${K8S_VERSION}/kubernetes-server-linux-amd64.tar.gz" \
    "kube-apiserver" \
    "$BIN_DIR"

download_and_install \
    "https://dl.k8s.io/${K8S_VERSION}/kubernetes-client-linux-amd64.tar.gz" \
    "kubectl" \
    "$BIN_DIR"

step "Настройка прав"
chmod +x "$BIN_DIR/"* || error "Ошибка установки прав доступа"
log "Установлены права выполнения для всех бинарников"

step "Завершение установки"
log "Бинарники установлены в: $BIN_DIR"
ls -la "$BIN_DIR"

echo -e "\n${GREEN}Инструкции:${NC}"
echo "Запуск тестов: make test-env"
echo "Использование kubectl: KUBECONFIG=./bin/kubeconfig.yaml ./bin/kubectl get kafkousers -A"
