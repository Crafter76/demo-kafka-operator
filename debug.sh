#!/bin/bash

set -euo pipefail

colors() {
    local -A cols=(
        [green]="\033[0;32m"
        [red]="\033[0;31m"
        [yellow]="\033[1;33m"
        [nc]="\033[0m"
    )
    echo "${cols[$1]}${2}${cols[nc]}"
}

log() { colors green "INFO $1"; }
error() { colors red "ERROR $1"; exit 1; }
warn() { colors yellow "WARN $1"; }
step() { echo -e "\n=== $1 ==="; }

check() {
    step "Проверка окружения"
    
    # Go и go.mod
    command -v go >/dev/null || error "Go не установлен"
    [[ -f go.mod ]] || error "go.mod не найден"
    log "Go: $(go version | awk '{print $3}')"
    log "Модуль: $(go list)"
    
    # Бинарники
    local bins=("etcd" "kube-apiserver" "kubectl")
    for bin in "${bins[@]}"; do
        local path="./bin/$bin"
        [[ -f "$path" ]] || error "$path не найден"
        [[ -x "$path" ]] || { chmod +x "$path"; log "Исправлено: $path теперь исполняемый"; }
        log "$bin: $(file "$path" | cut -d',' -f1)"
    done
    
    # DeepCopy и CRD
    [[ -f api/v1/zz_generated.deepcopy.go ]] || error "Файл deepcopy не найден"
    grep -q "DeepCopyObject" api/v1/zz_generated.deepcopy.go || error "Отсутствует DeepCopyObject"
    
    local crd="config/crd/kafkav1.crd.yaml"
    [[ -f "$crd" ]] || error "$crd не найден"
    grep -q "status: {}" "$crd" || error "Отсутствует subresources.status"
    grep -q "lastTransitionTime" "$crd" || error "Отсутствует lastTransitionTime"
    
    # Kubeconfig
    local kubeconfig="./bin/kubeconfig.yaml"
    if [[ -f "$kubeconfig" ]]; then
        if grep -q "server:" "$kubeconfig"; then
            export KUBECONFIG="$kubeconfig"
            log "kubeconfig корректен"
        else
            warn "kubeconfig может быть повреждён"
        fi
    else
        warn "kubeconfig не найден"
    fi
}

run_tests() {
    step "Запуск тестов"
    GOMAXPROCS=4 go test -v ./test/... -count=1 -p=1 -timeout 60s || {
        error "Тесты не прошли"
        echo
        echo "Возможные причины:"
        echo "• Отсутствует subresources.status в CRD"
        echo "• Неверный resourceVersion (используйте Patch())"
        echo "• Пользователь уже существует (проверяйте HasUser())"
        echo "• Отсутствует lastTransitionTime в схеме"
        exit 1
    }
}

main() {
    check
    run_tests
    log "Диагностика завершена"
}

main
