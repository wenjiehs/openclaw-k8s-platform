package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openclaw/openclaw-saas-platform/pkg/config"
	"github.com/openclaw/openclaw-saas-platform/pkg/logger"
)

// Client K8s 客户端封装
type Client struct {
	clientset *kubernetes.Clientset
	cfg       *config.TKEConfig
	log       *logger.Logger
}

// NewClient 创建 K8s 客户端
// 优先使用 kubeconfig，在 K8s 集群内时使用 in-cluster 配置
func NewClient(cfg *config.Config, log *logger.Logger) (*Client, error) {
	var k8sCfg *rest.Config
	var err error

	// 尝试使用 in-cluster 配置（在 K8s Pod 中运行时）
	k8sCfg, err = rest.InClusterConfig()
	if err != nil {
		// 退回到 kubeconfig 文件（本地开发时）
		k8sCfg, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("初始化 K8s 配置失败: %w", err)
		}
		log.Info("使用 kubeconfig 文件连接 K8s 集群")
	} else {
		log.Info("使用 in-cluster 配置连接 K8s 集群")
	}

	clientset, err := kubernetes.NewForConfig(k8sCfg)
	if err != nil {
		return nil, fmt.Errorf("创建 K8s 客户端失败: %w", err)
	}

	return &Client{
		clientset: clientset,
		cfg:       &cfg.TKE,
		log:       log,
	}, nil
}

// CreateNamespace 创建用户专属 Namespace
func (c *Client) CreateNamespace(ctx context.Context, namespace, username, department string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"app":        "openclaw",
				"user":       username,
				"department": department,
				"managed-by": "openclaw-saas",
			},
		},
	}

	_, err := c.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		c.log.Info(fmt.Sprintf("Namespace %s 已存在", namespace))
		return nil
	}
	return err
}

// DeleteNamespace 删除 Namespace（会级联删除所有资源）
func (c *Client) DeleteNamespace(ctx context.Context, namespace string) error {
	err := c.clientset.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		c.log.Info(fmt.Sprintf("Namespace %s 不存在，跳过删除", namespace))
		return nil
	}
	return err
}

// CreatePVC 创建持久化存储
func (c *Client) CreatePVC(ctx context.Context, namespace string, storageSize string) error {
	storageClass := "tke-cbs" // TKE 云硬盘存储类
	quantity, err := parseQuantity(storageSize)
	if err != nil {
		return err
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openclaw-data",
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: &storageClass,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: quantity,
				},
			},
		},
	}

	_, err = c.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// GetNamespaceStatus 获取 Namespace 中 OpenClaw Pod 的运行状态
func (c *Client) GetNamespaceStatus(ctx context.Context, namespace string) (string, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=openclaw",
	})
	if err != nil {
		return "unknown", err
	}

	if len(pods.Items) == 0 {
		return "pending", nil
	}

	pod := pods.Items[0]
	switch pod.Status.Phase {
	case corev1.PodRunning:
		return "running", nil
	case corev1.PodPending:
		return "creating", nil
	case corev1.PodFailed:
		return "failed", nil
	default:
		return "unknown", nil
	}
}

// CreateDeployment 创建 OpenClaw Deployment
func (c *Client) CreateDeployment(ctx context.Context, namespace, username, imageName string, resources ResourceRequirements) error {
	replicas := int32(1)

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openclaw",
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "openclaw"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "openclaw",
						"user": username,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "openclaw-all-in-one",
							Image: imageName,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080},
							},
							Resources: resources.ToK8sResources(),
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/app/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "openclaw-data",
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := c.clientset.AppsV1().Deployments(namespace).Create(ctx, deploy, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// CreateService 创建 ClusterIP Service
func (c *Client) CreateService(ctx context.Context, namespace string) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openclaw-service",
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "openclaw"},
			Ports: []corev1.ServicePort{
				{
					Port:     80,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err := c.clientset.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// CreateIngress 创建 Ingress（对外暴露服务）
func (c *Client) CreateIngress(ctx context.Context, namespace, hostname string) error {
	pathType := networkingv1.PathTypePrefix
	ingressClassName := "nginx"

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openclaw-ingress",
			Namespace: namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":      "nginx",
				"cert-manager.io/cluster-issuer":   "letsencrypt-prod",
				"nginx.ingress.kubernetes.io/proxy-connect-timeout": "60",
				"nginx.ingress.kubernetes.io/proxy-read-timeout":    "300",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClassName,
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{hostname},
					SecretName: namespace + "-tls",
				},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "openclaw-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := c.clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
