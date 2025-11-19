package k8s

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwardSession struct {
	PodName    string
	Namespace  string
	LocalPort  int
	RemotePort int
	StopChan   chan struct{}
	ReadyChan  chan struct{}
	ErrorChan  chan error
	client     kubernetes.Interface
	config     *rest.Config
}

func NewPortForwardSession(podName, namespace string, localPort, remotePort int, client kubernetes.Interface, config *rest.Config) *PortForwardSession {
	return &PortForwardSession{
		PodName:    podName,
		Namespace:  namespace,
		LocalPort:  localPort,
		RemotePort: remotePort,
		StopChan:   make(chan struct{}, 1),
		ReadyChan:  make(chan struct{}, 1),
		ErrorChan:  make(chan error, 1),
		client:     client,
		config:     config,
	}
}

func (pf *PortForwardSession) Start() error {
	req := pf.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(pf.Namespace).
		Name(pf.PodName).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(pf.config)
	if err != nil {
		return fmt.Errorf("failed to create round tripper: %v", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())

	ports := []string{fmt.Sprintf("%d:%d", pf.LocalPort, pf.RemotePort)}

	// Use a discard writer to prevent printing "Forwarding from..." messages
	discardWriter := io.Discard
	pfForwarder, err := portforward.New(dialer, ports, pf.StopChan, pf.ReadyChan, discardWriter, discardWriter)
	if err != nil {
		return fmt.Errorf("failed to create port forwarder: %v", err)
	}

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			close(pf.StopChan)
		case <-pf.StopChan:
		}
	}()

	go func() {
		err := pfForwarder.ForwardPorts()
		if err != nil {
			pf.ErrorChan <- fmt.Errorf("port forwarding failed: %v", err)
		}
	}()

	select {
	case <-pf.ReadyChan:
		return nil
	case err := <-pf.ErrorChan:
		return err
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for port forwarding to start")
	}
}

func (pf *PortForwardSession) Stop() {
	close(pf.StopChan)
}

func (p *Pod) PortForward(localPort, remotePort int) (*PortForwardSession, error) {
	if p.Raw == nil {
		if err := p.Fetch(); err != nil {
			return nil, fmt.Errorf("failed to fetch pod: %v", err)
		}
	}

	// Check if pod is running
	if p.Raw.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("pod is not running (status: %s)", p.Raw.Status.Phase)
	}

	session := NewPortForwardSession(p.Name, p.Namespace, localPort, remotePort, p.Client, p.Config)

	if err := session.Start(); err != nil {
		return nil, err
	}

	return session, nil
}

type PortForwardManager struct {
	sessions map[string]*PortForwardSession
}

func NewPortForwardManager() *PortForwardManager {
	return &PortForwardManager{
		sessions: make(map[string]*PortForwardSession),
	}
}

func (pm *PortForwardManager) AddSession(key string, session *PortForwardSession) {
	pm.sessions[key] = session
}

func (pm *PortForwardManager) RemoveSession(key string) {
	if session, exists := pm.sessions[key]; exists {
		session.Stop()
		delete(pm.sessions, key)
	}
}

func (pm *PortForwardManager) GetSession(key string) (*PortForwardSession, bool) {
	session, exists := pm.sessions[key]
	return session, exists
}

func (pm *PortForwardManager) ListSessions() map[string]*PortForwardSession {
	return pm.sessions
}

func (pm *PortForwardManager) StopAll() {
	for key, session := range pm.sessions {
		session.Stop()
		delete(pm.sessions, key)
	}
}

func (pm *PortForwardManager) GetSessionInfo() string {
	if len(pm.sessions) == 0 {
		return "No active port forwarding sessions"
	}

	type SessionInfo struct {
		Key        string `yaml:"key"`
		PodName    string `yaml:"podName"`
		Namespace  string `yaml:"namespace"`
		LocalPort  int    `yaml:"localPort"`
		RemotePort int    `yaml:"remotePort"`
	}

	var sessions []SessionInfo
	for key, session := range pm.sessions {
		sessions = append(sessions, SessionInfo{
			Key:        key,
			PodName:    session.PodName,
			Namespace:  session.Namespace,
			LocalPort:  session.LocalPort,
			RemotePort: session.RemotePort,
		})
	}

	// Format as YAML-like output
	output := "Active Port Forwarding Sessions:\n"
	for _, session := range sessions {
		output += fmt.Sprintf("- %s: %s/%s (localhost:%d -> %d)\n",
			session.Key, session.Namespace, session.PodName, session.LocalPort, session.RemotePort)
	}

	return output
}

// Helper function to parse port strings
func ParsePort(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", portStr)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port must be between 1 and 65535")
	}
	return port, nil
}

// Helper function to validate port availability
func IsPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// Helper function to display port forwarding status
func DisplayPortForwardStatus(podName, namespace string, localPort, remotePort int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2).
		BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	message := fmt.Sprintf("Port forwarding started!\n\n"+
		"Pod: %s/%s\n"+
		"Forwarding: localhost:%d -> %d\n\n"+
		"Press Ctrl+C to stop port forwarding",
		namespace, podName, localPort, remotePort)

	return style.Render(message)
}
