package kube

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"

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
