
# Скачиваем etcd (v3.5.9) для linux/amd64
curl -L https://github.com/etcd-io/etcd/releases/download/v3.6.5/etcd-v3.6.5-linux-amd64.tar.gz -o etcd.zip
unzip etcd.zip etcd-v3.6.5-linux-amd64/etcd
mv etcd-v3.6.5-linux-amd64/etcd demo-kafka-operator/bin/etcd

# Скачиваем kube-apiserver (v1.34.1) для linux/amd64
curl -L https://dl.k8s.io/v1.34.1/kubernetes-server-linux-amd64.tar.gz | tar -zx kubernetes/server/bin/kube-apiserver
mv kubernetes/server/bin/kube-apiserver demo-kafka-operator/bin/kube-apiserver

# Делаем исполняемыми
chmod +x demo-kafka-operator/bin/etcd
chmod +x demo-kafka-operator/bin/kube-apiserver