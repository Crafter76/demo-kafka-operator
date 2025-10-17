package test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	v1 "github.com/crafter76/demo-kafka-operator/api/v1"
	"github.com/crafter76/demo-kafka-operator/controller"
	"github.com/crafter76/demo-kafka-operator/mocks"
)

var (
	testEnv    *envtest.Environment
	ctx        context.Context
	cancel     context.CancelFunc
	k8sClient  client.Client
	k8sManager ctrl.Manager
	mockKafka  *mocks.MockKafkaClient
	scheme     = runtime.NewScheme()
)

func ptrTo[T any](v T) *T {
	return &v
}

func TestMain(m *testing.M) {
	// Настройка scheme
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1.AddToScheme(scheme))

	// Настройка envtest
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:        []string{filepath.Join("..", "config", "crd")},
		Scheme:                   scheme,
		UseExistingCluster:       ptrTo(false),
		BinaryAssetsDirectory:    filepath.Join("..", "bin"),
		ControlPlaneStartTimeout: 3 * time.Minute,
	}

	// Запуск control plane
	cfg, err := testEnv.Start()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := testEnv.Stop(); err != nil {
			panic(err)
		}
	}()

	// Создание manager
	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme:         scheme,
		Metrics:        server.Options{BindAddress: "0"},
		LeaderElection: false, // выключаем для теста
	})
	if err != nil {
		panic(err)
	}

	// Включаем логирование (чтобы видеть reconcile)
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Инициализируем mock
	mockKafka = mocks.NewMockKafkaClient()

	// Регистрируем контроллер
	reconciler := &controller.KafkaUserReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
		Kafka:  mockKafka,
	}
	if err := reconciler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}

	// Сохраняем клиент
	k8sClient = k8sManager.GetClient()

	// Запускаем manager в фоне
	ctx, cancel = context.WithCancel(context.TODO())
	go func() {
		if err := k8sManager.Start(ctx); err != nil {
			panic(err)
		}
	}()

	// Запускаем тесты
	m.Run()

	// Останавливаем
	cancel()

	<-k8sManager.Elected()
}

func TestKafkaUserCreation(t *testing.T) {
	g := NewWithT(t)

	t.Log("1. Создаём CR KafkaUser")
	user := &v1.KafkaUser{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-user", Namespace: "default"},
		Spec: v1.KafkaUserSpec{
			Topic:       "payments",
			Permissions: "read",
		},
	}
	g.Expect(k8sClient.Create(ctx, user)).To(Succeed())

	t.Log("2. Проверяем, что CR действительно создан")
	g.Eventually(func() error {
		return k8sClient.Get(ctx, client.ObjectKeyFromObject(user), &v1.KafkaUser{})
	}, 5*time.Second, time.Second).Should(Succeed())

	t.Log("3. Проверяем, что reconcile запустился (mock Kafka был вызван)")
	g.Eventually(func() int {
		return mockKafka.Calls // счётчик из mocks/mock_kafka.go
	}, 5*time.Second, time.Second).Should(BeNumerically(">", 0), "Ожидался вызов mockKafka.CreateUser")

	t.Log("4. Проверяем, что finalizer добавлен")
	g.Eventually(func() []string {
		var u v1.KafkaUser
		_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(user), &u)
		return u.Finalizers
	}, 10*time.Second, time.Second).Should(ContainElement("kafka.appfarm.rs/finalizer"))

	t.Log("5. Проверяем, что статус обновлён: Created=True")
	g.Eventually(func() []metav1.Condition {
		var u v1.KafkaUser
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(user), &u); err != nil {
			return nil
		}
		return u.Status.Conditions
	}, 5*time.Second, time.Second).Should(
		ContainElement(And(
			HaveField("Type", "Created"),
			HaveField("Status", metav1.ConditionTrue),
			HaveField("Reason", "Success"),
			HaveField("Message", "User created in Kafka"),
		)),
		"Контроллер должен установить статус Created=True после успешного создания пользователя",
	)

	t.Log("6. Финальная проверка: пользователь есть в mock")
	g.Expect(mockKafka.HasUser("dev-user")).To(BeTrue(), "mockKafka должен содержать пользователя 'dev-user'")
}
