package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openclaw/openclaw-saas-platform/internal/model"
)

// ResourceRequirements K8s 资源需求
type ResourceRequirements struct {
	CPURequest    string
	MemoryRequest string
	CPULimit      string
	MemoryLimit   string
}

// ToK8sResources 转换为 K8s ResourceRequirements 格式
func (r ResourceRequirements) ToK8sResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(r.CPURequest),
			corev1.ResourceMemory: resource.MustParse(r.MemoryRequest),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(r.CPULimit),
			corev1.ResourceMemory: resource.MustParse(r.MemoryLimit),
		},
	}
}

// parseQuantity 解析资源数量字符串
func parseQuantity(s string) (resource.Quantity, error) {
	return resource.ParseQuantity(s)
}

// InstanceDeployConfig 实例部署配置
type InstanceDeployConfig struct {
	Username   string
	Namespace  string
	Spec       string
	Department string
	Domain     string     // 实例域名后缀，如 tke-cloud.com
	ImageName  string     // OpenClaw 镜像名称
}

// Operator OpenClaw 实例编排器
// 负责完整的实例生命周期管理（创建/删除）
type Operator struct {
	k8sClient *Client
}

// NewOperator 创建实例编排器
func NewOperator(k8sClient *Client) *Operator {
	return &Operator{k8sClient: k8sClient}
}

// CreateInstance 创建完整的 OpenClaw 实例
// 按顺序执行：Namespace -> PVC -> Deployment -> Service -> Ingress
func (o *Operator) CreateInstance(ctx context.Context, cfg InstanceDeployConfig) (accessURL string, err error) {
	log := o.k8sClient.log
	namespace := cfg.Namespace
	hostname := fmt.Sprintf("%s.%s", cfg.Username, cfg.Domain)

	log.Info(fmt.Sprintf("开始创建实例: %s, namespace: %s", cfg.Username, namespace))

	// 根据规格获取资源配置
	specConfig := model.GetResourceConfig(cfg.Spec)
	resources := ResourceRequirements{
		CPURequest:    specConfig.CPU,
		MemoryRequest: specConfig.Memory,
		CPULimit:      specConfig.CPU,
		MemoryLimit:   specConfig.Memory,
	}

	// 步骤 1：创建 Namespace
	log.Info(fmt.Sprintf("步骤 1/5: 创建 Namespace %s", namespace))
	if err := o.k8sClient.CreateNamespace(ctx, namespace, cfg.Username, cfg.Department); err != nil {
		return "", fmt.Errorf("创建 Namespace 失败: %w", err)
	}

	// 步骤 2：创建 PVC（10GB 存储）
	log.Info(fmt.Sprintf("步骤 2/5: 创建 PVC，存储大小: 10Gi"))
	if err := o.k8sClient.CreatePVC(ctx, namespace, "10Gi"); err != nil {
		return "", fmt.Errorf("创建 PVC 失败: %w", err)
	}

	// 步骤 3：创建 Deployment
	imageName := cfg.ImageName
	if imageName == "" {
		imageName = "registry.tke-cloud.com/openclaw/openclaw:latest"
	}
	log.Info(fmt.Sprintf("步骤 3/5: 创建 Deployment，镜像: %s", imageName))
	if err := o.k8sClient.CreateDeployment(ctx, namespace, cfg.Username, imageName, resources); err != nil {
		return "", fmt.Errorf("创建 Deployment 失败: %w", err)
	}

	// 步骤 4：创建 Service
	log.Info("步骤 4/5: 创建 Service")
	if err := o.k8sClient.CreateService(ctx, namespace); err != nil {
		return "", fmt.Errorf("创建 Service 失败: %w", err)
	}

	// 步骤 5：创建 Ingress
	log.Info(fmt.Sprintf("步骤 5/5: 创建 Ingress，域名: %s", hostname))
	if err := o.k8sClient.CreateIngress(ctx, namespace, hostname); err != nil {
		return "", fmt.Errorf("创建 Ingress 失败: %w", err)
	}

	accessURL = fmt.Sprintf("https://%s", hostname)
	log.Info(fmt.Sprintf("实例创建完成: %s, 访问地址: %s", cfg.Username, accessURL))

	return accessURL, nil
}

// WaitForInstanceReady 等待实例就绪（Pod Running 状态）
// timeout: 超时时间
func (o *Operator) WaitForInstanceReady(ctx context.Context, namespace string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second) // 每 10 秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context 已取消")
		case <-ticker.C:
			status, err := o.k8sClient.GetNamespaceStatus(ctx, namespace)
			if err != nil {
				o.k8sClient.log.Warn(fmt.Sprintf("检查实例状态失败: %v", err))
				continue
			}

			if status == "running" {
				return nil // 实例就绪
			}

			if status == "failed" {
				return fmt.Errorf("实例启动失败")
			}

			if time.Now().After(deadline) {
				return fmt.Errorf("等待实例就绪超时（%v）", timeout)
			}

			o.k8sClient.log.Info(fmt.Sprintf("实例状态: %s，继续等待...", status))
		}
	}
}

// DeleteInstance 删除实例的所有 K8s 资源
// 只需删除 Namespace，会级联删除所有资源
func (o *Operator) DeleteInstance(ctx context.Context, namespace string) error {
	o.k8sClient.log.Info(fmt.Sprintf("开始删除实例: namespace=%s", namespace))
	if err := o.k8sClient.DeleteNamespace(ctx, namespace); err != nil {
		return fmt.Errorf("删除 Namespace 失败: %w", err)
	}
	o.k8sClient.log.Info(fmt.Sprintf("实例删除成功: namespace=%s", namespace))
	return nil
}

// GetInstanceMetrics 获取实例的资源使用指标
// TODO: 从 Prometheus 获取真实数据，目前返回模拟数据
func (o *Operator) GetInstanceMetrics(ctx context.Context, namespace string) (*InstanceMetrics, error) {
	// 获取 Pod 列表
	pods, err := o.k8sClient.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=openclaw",
	})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return &InstanceMetrics{Status: "stopped"}, nil
	}

	pod := pods.Items[0]
	metrics := &InstanceMetrics{
		PodName:   pod.Name,
		Status:    string(pod.Status.Phase),
		StartTime: pod.Status.StartTime,
	}

	return metrics, nil
}

// InstanceMetrics 实例监控指标
type InstanceMetrics struct {
	PodName     string
	Status      string
	CPUUsage    float64    // CPU 使用率（0-100%）
	MemoryUsage float64    // 内存使用率（0-100%）
	StartTime   *metav1.Time
}
