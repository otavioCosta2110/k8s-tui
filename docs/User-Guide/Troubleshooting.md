# Troubleshooting Guide

This guide helps you resolve common issues when using k8s-tui.

## Connection Issues

### Cannot Connect to Cluster

**Symptoms:**
- "connection refused" errors
- "unable to connect" messages
- Blank resource lists

**Solutions:**

1. **Check cluster connectivity:**
   ```bash
   kubectl cluster-info
   ```

2. **Verify kubeconfig:**
   ```bash
   kubectl config view
   kubectl config current-context
   ```

3. **Test API access:**
   ```bash
   kubectl get nodes
   kubectl get pods --all-namespaces
   ```

4. **Check network connectivity:**
   ```bash
   curl -k https://<api-server>:6443/api/v1/nodes
   ```

### Authentication Errors

**Symptoms:**
- "Unauthorized" or "Forbidden" errors
- Permission denied messages

**Solutions:**

1. **Check RBAC permissions:**
   ```bash
   kubectl auth can-i list pods
   kubectl auth can-i create deployments
   ```

2. **Verify token/certificate validity:**
   ```bash
   kubectl config view --minify
   # Check expiration dates
   ```

3. **Refresh authentication:**
   ```bash
   kubectl config unset users.<user>
   # Re-login or refresh credentials
   ```

## Resource Management Issues

### Cannot Create Resources

**Symptoms:**
- Form submission fails
- "insufficient permissions" errors

**Solutions:**

1. **Check namespace permissions:**
   ```bash
   kubectl auth can-i create pods -n <namespace>
   ```

2. **Verify resource quotas:**
   ```bash
   kubectl get resourcequotas -n <namespace>
   kubectl describe resourcequota <quota-name> -n <namespace>
   ```

3. **Check cluster resources:**
   ```bash
   kubectl get nodes --show-capacity
   ```

### Cannot Edit Resources

**Symptoms:**
- Editor doesn't open
- Changes not applied

**Solutions:**

1. **Check EDITOR environment variable:**
   ```bash
   echo $EDITOR
   export EDITOR=vim  # or nano, code, etc.
   ```

2. **Verify editor installation:**
   ```bash
   which $EDITOR
   ```

3. **Check file permissions:**
   ```bash
   # Ensure temp directory is writable
   mkdir -p /tmp/k8s-tui
   ```

## Display Issues

### Garbled Text / Wrong Colors

**Symptoms:**
- Icons not displaying
- Colors not rendering
- Layout broken

**Solutions:**

1. **Check terminal capabilities:**
   ```bash
   echo $TERM
   # Should be xterm-256color or similar
   ```

2. **Enable Unicode support:**
   ```bash
   export LANG=en_US.UTF-8
   export LC_ALL=en_US.UTF-8
   ```

3. **Reset theme:**
   ```bash
   cd assets/colorschemes
   cp one-dark.json current-theme.json
   ```

### Screen Size Issues

**Symptoms:**
- Content cut off
- Layout overlapping

**Solutions:**

1. **Minimum terminal size:** 120x30 characters
2. **Check window size:**
   ```bash
   stty size
   # Should show at least 30 rows, 120 columns
   ```

3. **Resize terminal and restart k8s-tui**

## Performance Issues

### Slow Loading

**Symptoms:**
- Long load times for resource lists
- Delayed updates

**Solutions:**

1. **Check cluster size:**
   ```bash
   kubectl get nodes
   kubectl get pods --all-namespaces | wc -l
   ```

2. **Reduce refresh interval:**
   ```bash
   # Currently not configurable, but check network latency
   ping <api-server>
   ```

3. **Use namespace filtering:**
   ```bash
   # Instead of --all-namespaces, specify namespace
   kubectl config set-context --current --namespace=<namespace>
   ```

### High Memory Usage

**Symptoms:**
- Application becomes slow
- System memory usage increases

**Solutions:**

1. **Close unused tabs**
2. **Reduce concurrent operations**
3. **Check for memory leaks in plugins**

## Plugin Issues

### Plugin Not Loading

**Symptoms:**
- Plugin doesn't appear in interface
- Error messages on startup

**Solutions:**

1. **Check plugin syntax:**
   ```bash
   lua -l plugins/<plugin>.lua
   ```

2. **Verify plugin directory:**
   ```bash
   ls -la plugins/
   ```

3. **Check plugin logs:**
   ```bash
   tail -f ~/.k8s-tui/logs/plugin.log
   ```

### Plugin Errors

**Symptoms:**
- Plugin crashes
- Unexpected behavior

**Solutions:**

1. **Update plugin to latest version**
2. **Check compatibility with k8s-tui version**
3. **Report issues to plugin maintainer**

## Kubernetes-Specific Issues

### Resource Quota Exceeded

**Symptoms:**
- Cannot create resources
- "exceeded quota" errors

**Solutions:**

1. **Check current usage:**
   ```bash
   kubectl get resourcequotas -n <namespace>
   kubectl describe resourcequota <quota> -n <namespace>
   ```

2. **Clean up unused resources:**
   ```bash
   kubectl delete pods --field-selector=status.phase=Succeeded
   ```

3. **Request quota increase from cluster admin**

### Image Pull Issues

**Symptoms:**
- Pods stuck in "ImagePullBackOff" status

**Solutions:**

1. **Check image name and tag:**
   ```bash
   kubectl describe pod <pod-name>
   # Look for image pull errors
   ```

2. **Verify registry access:**
   ```bash
   docker pull <image>  # if using Docker
   ```

3. **Check image pull secrets:**
   ```bash
   kubectl get secrets -n <namespace>
   ```

### Network Issues

**Symptoms:**
- Services not accessible
- Pod-to-pod communication fails

**Solutions:**

1. **Check service configuration:**
   ```bash
   kubectl get svc <service>
   kubectl describe svc <service>
   ```

2. **Verify network policies:**
   ```bash
   kubectl get networkpolicies -n <namespace>
   ```

3. **Check DNS resolution:**
   ```bash
   kubectl exec -it <pod> -- nslookup <service>
   ```

## Logs and Debugging

### Enable Debug Logging

```bash
export K8S_TUI_LOG_LEVEL=debug
./k8s-tui
```

### View Application Logs

```bash
tail -f ~/.k8s-tui/logs/app.log
```

### Kubernetes Logs

```bash
# View k8s-tui pod logs (if running in cluster)
kubectl logs -f deployment/k8s-tui

# View API server logs
kubectl logs -n kube-system kube-apiserver-<node>
```

## Getting Help

### Community Support

- [GitHub Issues](https://github.com/otavioCosta2110/k8s-tui/issues)
- [GitHub Discussions](https://github.com/otavioCosta2110/k8s-tui/discussions)

### Reporting Bugs

When reporting issues, include:

1. **k8s-tui version:**
   ```bash
   ./k8s-tui --version
   ```

2. **Go version:**
   ```bash
   go version
   ```

3. **Kubernetes version:**
   ```bash
   kubectl version
   ```

4. **Operating system and terminal:**
   ```bash
   uname -a
   echo $TERM
   ```

5. **Steps to reproduce**
6. **Expected vs actual behavior**
7. **Relevant logs**

### Feature Requests

Use GitHub Issues with the "enhancement" label to request new features.</content>
