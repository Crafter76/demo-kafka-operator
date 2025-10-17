#!/bin/bash
# debug.sh — диагностика env-тестов для KafkaUserReconciler

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

step() {
    echo
    echo -e "${YELLOW}=== $1 ===${NC}"
}

# Проверка Go
step "Проверка Go"
if ! command -v go &> /dev/null; then
    error "Go не установлен"
    exit 1
fi
log "Go version: $(go version)"

# Проверка модуля
step "Проверка go.mod"
if [ ! -f go.mod ]; then
    error "go.mod не найден"
    exit 1
fi
log "Модуль: $(go list -m)"
log "Зависимости controller-runtime:"
go list -m sigs.k8s.io/controller-runtime || true

# Проверка bin/
step "Проверка bin/"
if [ ! -d bin ]; then
    error "Папка bin/ не существует"
    exit 1
fi

for binary in etcd kube-apiserver; do
    if [ ! -f "bin/$binary" ]; then
        error "bin/$binary не найден"
        exit 1
    fi
    if [ ! -x "bin/$binary" ]; then
        chmod +x "bin/$binary" && log "Исправлено: bin/$binary теперь исполняемый"
    fi
    log "$binary: $(file bin/$binary | cut -d',' -f1)"
done

# Проверка deepcopy
step "Проверка deepcopy"
if [ ! -f api/v1/zz_generated.deepcopy.go ]; then
    warn "Нет zz_generated.deepcopy.go — генерируем..."
    go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0 object paths="./..." || { error "Генерация deepcopy не удалась"; exit 1; }
else
    log "deepcopy.go найден"
    grep -q "DeepCopyObject" api/v1/zz_generated.deepcopy.go && log "✅ DeepCopyObject реализован" || error "❌ Нет DeepCopyObject"
fi

# Проверка CRD
step "Проверка CRD"
if [ ! -f config/crd/kafkav1.crd.yaml ]; then
    error "CRD не найден"
    exit 1
fi
log "CRD найден"

# Запуск теста с подробными логами
step "Запуск теста"
GOMAXPROCS=4 go test -v ./test/... -count=1 -p=1 -timeout 30s

echo
log "Диагностика завершена."