# KafkaUser Operator Demo

Демонстрация оператора Kubernetes для управления пользователями Kafka.  
Используется в докладе на [Гейзенбаге](https://heisenbug.ru).

> ✅ Работает **без Docker, kind, minikube**  
> ✅ Не требует интернета после подготовки  
> ✅ Использует настоящие `etcd`, `kube-apiserver`, `kubectl`  
> ✅ Полностью автономный env-тест  

---

## 🎯 Цель

Показать, как можно:
- Тестировать оператор **как обычный unit-тест**
- Запускать временный control plane локально
- Обеспечивать надёжность reconcile без инфраструктуры

---

## 🧩 Структура проекта

```
demo-kafka-operator/
├── api/v1/                  # CRD: KafkaUser
├── controller/              # Reconciler
├── mocks/                   # Mock Kafka API
├── test/                    # Env-тесты
├── config/crd/              # kafkav1.crd.yaml
├── go.mod
├── Makefile
├── install.sh               # Установка бинарников
└── debug.sh                 # Диагностика и запуск тестов
```

---

## ⚙️ Зависимости

- Go 1.21+
- Linux (amd64)
- bash, curl, tar, unzip

---

## 📦 Установка бинарников

Если папка `bin/` пуста или повреждена:

```bash
./install.sh
```

Скрипт скачает:
- `etcd`
- `kube-apiserver`
- `kubectl`
— и разместит их в `./bin/`

---

## ▶️ Запуск тестов

```bash
make test-env
```

Запускает env-тест:
- Временный control plane (`etcd` + `kube-apiserver`)
- Создание CR `KafkaUser`
- Проверка reconcile, статуса и mock-взаимодействия
- Автоматическое сохранение `./bin/kubeconfig.yaml`

---

## 🛠️ Диагностика

Если тест не проходит — используйте скрипт диагностики:

```bash
./debug.sh
```

Он проверит:
- Наличие `go`, `etcd`, `kube-apiserver`, `kubectl`
- Корректность `deepcopy.go`
- Соответствие CRD (`subresources.status`, `lastTransitionTime`)
- Запустит тест с подробными логами
- Попробует показать состояние через `kubectl`

---


## 🧪 Как работает env-тест

1. `envtest` запускает `etcd` и `kube-apiserver` из `./bin`
2. Контроллер регистрируется и начинает обрабатывать события
3. Тест создаёт `KafkaUser`
4. Reconcile:
   - Добавляет finalizer
   - Создаёт пользователя в mock-Kafka
   - Обновляет статус через `Status().Patch()`
5. Тест проверяет результат через `Eventually`
6. `kubeconfig.yaml` сохраняется для внешнего использования

---

## ❗ Распространённые проблемы

### `Failed to update status: not found`
- Причина: stale `resourceVersion` после `Update()`
- Решение: всегда используйте `client.MergeFrom(original)` + `Patch()`

### `user already exists`
- Причина: повторный вызов `CreateUser` после сбоя статуса
- Решение: проверяйте `HasUser()` перед созданием

### `lastTransitionTime` не проходит валидацию
- Причина: поле отсутствует в OpenAPI-схеме CRD
- Решение: явно объявите `lastTransitionTime` в `kafkav1.crd.yaml`


## 📦 Автор

[Василий Кирнос](https://t.me/crafter76)  
RSHB | Heisenbug 2025
