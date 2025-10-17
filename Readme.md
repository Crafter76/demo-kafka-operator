# KafkaUser Operator Demo

Демонстрация оператора Kubernetes для управления пользователями Kafka.  
Используется в докладе на [Гейзенбаге](https://heisenbug.ru).

> ✅ Работает **без Docker, kind, minikube**  
> ✅ Не требует интернета после подготовки  
> ✅ Использует настоящие `etcd` и `kube-apiserver`  
> ✅ Полностью автономный env-тест

---

## 🎯 Цель

Показать, как можно:
- Тестировать оператор **как unit-тест**
- Запускать `kube-apiserver` локально
- Управлять состоянием через `kubectl`
- Обеспечивать надёжность reconcile

---

## 🧩 Структура проекта

```
demo-kafka-operator/
├── api/v1/                  # CRD: KafkaUser
├── controller/              # Reconciler
├── mocks/                   # Mock Kafka API
├── test/                    # Env-тесты
├── config/crd/              # kafkav1.crd.yaml
├── bin/                     # etcd, kube-apiserver, kubectl
├── go.mod
└── Makefile
```

---

## ⚙️ Зависимости

- Go 1.21+
- Linux (amd64)

---

## 🔽 Бинарники в `bin/`

Проект содержит предустановленные бинарники:

| Бинарник | Версия | Откуда |
|--------|-------|--------|
| `etcd` | v3.5.9 | [GitHub](https://github.com/etcd-io/etcd/releases/tag/v3.5.9) |
| `kube-apiserver` | v1.27.0 | [Kubernetes Release](https://dl.k8s.io/v1.27.0/bin/linux/amd64/kube-apiserver) |

> 💡 Все бинарники скачаны из официальных источников.

---

## ▶️ Запуск тестов

```bash
make test-env
```

Запускает env-тест:
- Временный control plane (`etcd` + `kube-apiserver`)
- Создание CR `KafkaUser`
- Проверка reconcile, статуса и mock-взаимодействия

---

## 🛠️ Диагностика

Если тест не проходит — используйте скрипт диагностики:

```bash
./debug.sh
```

Он проверит:
- Наличие `go`, `etcd`, `kube-apiserver`
- Корректность `deepcopy.go`
- Соответствие CRD
- И запустит тест с подробными логами

---

## 🧪 Как работает env-тест

1. `envtest` запускает `etcd` и `kube-apiserver` из `./bin`
2. Контроллер регистрируется и начинает обрабатывать события
3. Тест создаёт `KafkaUser`
4. Reconcile:
   - Добавляет finalizer
   - Создаёт пользователя в mock-Kafka
   - Обновляет статус
5. Тест проверяет результат через `Eventually`


## 📦 Автор

[Василий Кирнос](https://t.me/crafter76)  
RSHB | Heisenbug 2025