package kube

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type Client struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
}

func New(kubeconfig string) (*Client, error) {
	var cfg *rest.Config
	var err error
	if strings.TrimSpace(kubeconfig) == "" {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", filepath.Clean(kubeconfig))
	}
	if err != nil {
		return nil, err
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{Clientset: cs, Config: cfg}, nil
}

func (c *Client) ListNodes(ctx context.Context) ([]map[string]interface{}, error) {
	nodes, err := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(nodes.Items))
	for _, n := range nodes.Items {
		status := "Unknown"
		for _, cond := range n.Status.Conditions {
			if cond.Type == corev1.NodeReady {
				status = string(cond.Status)
			}
		}
		out = append(out, map[string]interface{}{"name": n.Name, "status": status, "labels": n.Labels})
	}
	return out, nil
}

func (c *Client) ListPods(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}
	pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(pods.Items))
	for _, p := range pods.Items {
		out = append(out, map[string]interface{}{"namespace": p.Namespace, "name": p.Name, "status": string(p.Status.Phase), "restarts": restarts(p)})
	}
	return out, nil
}

func (c *Client) ListNamespaces(ctx context.Context) ([]map[string]interface{}, error) {
	nss, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(nss.Items))
	for _, ns := range nss.Items {
		out = append(out, map[string]interface{}{
			"name":   ns.Name,
			"status": string(ns.Status.Phase),
		})
	}
	return out, nil
}

func (c *Client) ListDeployments(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}
	deps, err := c.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(deps.Items))
	for _, d := range deps.Items {
		out = append(out, deploymentSummary(d))
	}
	return out, nil
}

func (c *Client) ListResource(ctx context.Context, kind, namespace string) ([]map[string]interface{}, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}
	switch strings.ToLower(kind) {
	case "pods":
		return c.ListPods(ctx, namespace)
	case "deployments":
		return c.ListDeployments(ctx, namespace)
	case "statefulsets":
		return c.listStatefulSets(ctx, namespace)
	case "daemonsets":
		return c.listDaemonSets(ctx, namespace)
	case "replicasets":
		return c.listReplicaSets(ctx, namespace)
	case "jobs":
		return c.listJobs(ctx, namespace)
	case "cronjobs":
		return c.listCronJobs(ctx, namespace)
	case "services":
		return c.listServices(ctx, namespace)
	case "ingresses":
		return c.listIngresses(ctx, namespace)
	case "configmaps":
		return c.listConfigMaps(ctx, namespace)
	case "secrets":
		return c.listSecrets(ctx, namespace)
	case "persistentvolumeclaims":
		return c.listPVCs(ctx, namespace)
	case "persistentvolumes":
		return c.listPVs(ctx)
	case "storageclasses":
		return c.listStorageClasses(ctx)
	case "namespaces":
		return c.ListNamespaces(ctx)
	case "events":
		return c.listEvents(ctx, namespace)
	case "nodes":
		return c.ListNodes(ctx)
	default:
		return nil, fmt.Errorf("unsupported resource kind: %s", kind)
	}
}

func (c *Client) PodLogs(ctx context.Context, namespace, pod, container string, tail int64) ([]string, error) {
	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(pod, &corev1.PodLogOptions{Container: container, TailLines: &tail})
	r, err := req.Stream(ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return []string{}, nil
	}
	return strings.Split(strings.TrimSpace(string(b)), "\n"), nil
}

func (c *Client) StreamPodLogs(ctx context.Context, namespace, pod, container string, tail int64) (io.ReadCloser, error) {
	follow := true
	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(pod, &corev1.PodLogOptions{Container: container, TailLines: &tail, Follow: follow})
	return req.Stream(ctx)
}

func (c *Client) ExecOnce(ctx context.Context, namespace, pod, container string, command []string) (string, string, error) {
	req := c.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec")
	req.VersionedParams(&corev1.PodExecOptions{Container: container, Command: command, Stdin: false, Stdout: true, Stderr: true, TTY: false}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.Config, "POST", req.URL())
	if err != nil {
		return "", "", err
	}
	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{Stdout: &stdout, Stderr: &stderr})
	return stdout.String(), stderr.String(), err
}

func ScanLines(r io.Reader, fn func(string) error) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		if err := fn(s.Text()); err != nil {
			return err
		}
	}
	return s.Err()
}

func restarts(p corev1.Pod) int32 {
	var total int32
	for _, cs := range p.Status.ContainerStatuses {
		total += cs.RestartCount
	}
	return total
}

func deploymentSummary(d appsv1.Deployment) map[string]interface{} {
	ready := int32(0)
	if d.Status.ReadyReplicas > 0 {
		ready = d.Status.ReadyReplicas
	}
	return map[string]interface{}{
		"namespace": d.Namespace,
		"name":      d.Name,
		"replicas":  d.Status.Replicas,
		"ready":     ready,
		"updated":   d.Status.UpdatedReplicas,
	}
}

func (c *Client) listStatefulSets(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	items, err := c.Clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"namespace": x.Namespace, "name": x.Name, "replicas": x.Status.Replicas, "ready": x.Status.ReadyReplicas}) }
	return out, nil
}
func (c *Client) listDaemonSets(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	items, err := c.Clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"namespace": x.Namespace, "name": x.Name, "desired": x.Status.DesiredNumberScheduled, "ready": x.Status.NumberReady}) }
	return out, nil
}
func (c *Client) listReplicaSets(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	items, err := c.Clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"namespace": x.Namespace, "name": x.Name, "replicas": x.Status.Replicas, "ready": x.Status.ReadyReplicas}) }
	return out, nil
}
func (c *Client) listJobs(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	items, err := c.Clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"namespace": x.Namespace, "name": x.Name, "succeeded": x.Status.Succeeded, "failed": x.Status.Failed}) }
	return out, nil
}
func (c *Client) listCronJobs(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	items, err := c.Clientset.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"namespace": x.Namespace, "name": x.Name, "schedule": x.Spec.Schedule, "suspend": x.Spec.Suspend != nil && *x.Spec.Suspend}) }
	return out, nil
}
func (c *Client) listServices(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	items, err := c.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"namespace": x.Namespace, "name": x.Name, "type": x.Spec.Type, "clusterIP": x.Spec.ClusterIP}) }
	return out, nil
}
func (c *Client) listIngresses(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	items, err := c.Clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items {
		host := ""
		if len(x.Spec.Rules) > 0 { host = x.Spec.Rules[0].Host }
		out = append(out, map[string]interface{}{"namespace": x.Namespace, "name": x.Name, "host": host})
	}
	return out, nil
}
func (c *Client) listConfigMaps(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	items, err := c.Clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"namespace": x.Namespace, "name": x.Name, "dataItems": len(x.Data)}) }
	return out, nil
}
func (c *Client) listSecrets(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	items, err := c.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"namespace": x.Namespace, "name": x.Name, "type": x.Type}) }
	return out, nil
}
func (c *Client) listPVCs(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	items, err := c.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"namespace": x.Namespace, "name": x.Name, "status": x.Status.Phase, "volume": x.Spec.VolumeName}) }
	return out, nil
}
func (c *Client) listPVs(ctx context.Context) ([]map[string]interface{}, error) {
	items, err := c.Clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"name": x.Name, "status": x.Status.Phase, "claim": x.Spec.ClaimRef}) }
	return out, nil
}
func (c *Client) listStorageClasses(ctx context.Context) ([]map[string]interface{}, error) {
	items, err := c.Clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"name": x.Name, "provisioner": x.Provisioner}) }
	return out, nil
}
func (c *Client) listEvents(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	items, err := c.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil { return nil, err }
	out := make([]map[string]interface{}, 0, len(items.Items))
	for _, x := range items.Items { out = append(out, map[string]interface{}{"namespace": x.Namespace, "name": x.Name, "reason": x.Reason, "type": x.Type, "message": x.Message}) }
	return out, nil
}
