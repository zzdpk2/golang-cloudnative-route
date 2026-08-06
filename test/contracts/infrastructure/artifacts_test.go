package infrastructure

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

type object = map[string]any

func TestKubernetesBase(t *testing.T) {
	objects := render(t, "kubectl", "kustomize", "deploy/inference-lab/overlays/learning")
	router := requireObject(t, objects, "Deployment", "inference-router")
	requireNumberAtLeast(t, router, 2, "spec", "replicas")
	requireProductionDeployment(t, router)
	requireSelectorContract(t, router)
	podSpec := requireMap(t, router, "spec", "template", "spec")
	requirePositiveNumber(t, podSpec, "terminationGracePeriodSeconds")
	if _, topology := podSpec["topologySpreadConstraints"]; !topology {
		if _, affinity := podSpec["affinity"]; !affinity {
			t.Fatal("Router Pod must state a multi-node placement policy")
		}
	}
	requireServiceContract(t, requireObject(t, objects, "Service", "inference-router"))
	for _, model := range []string{"model-a", "model-b"} {
		deployment := requireObject(t, objects, "Deployment", model)
		requireProductionDeployment(t, deployment)
		requireSelectorContract(t, deployment)
		requireServiceContract(t, requireObject(t, objects, "Service", model))
	}
	pdb := requireObject(t, objects, "PodDisruptionBudget", "inference-router")
	requireOneOf(t, requireMap(t, pdb, "spec"), "minAvailable", "maxUnavailable")
	hpa := requireObject(t, objects, "HorizontalPodAutoscaler", "inference-router")
	requirePositiveNumber(t, requireMap(t, hpa, "spec"), "minReplicas")
	requirePositiveNumber(t, requireMap(t, hpa, "spec"), "maxReplicas")
	requireNonEmptySlice(t, requireMap(t, hpa, "spec"), "metrics")
	network := requireObject(t, objects, "NetworkPolicy", "inference-router-default-deny")
	networkSpec := requireMap(t, network, "spec")
	requireNonEmptySlice(t, networkSpec, "policyTypes")
	for _, direction := range []string{"ingress", "egress"} {
		if peers, exists := networkSpec[direction]; exists {
			if list, ok := peers.([]any); !ok || len(list) != 0 {
				t.Fatalf("default-deny %s must grant no peers", direction)
			}
		}
	}
	allow := requireObject(t, objects, "NetworkPolicy", "inference-router-allow")
	allowSpec := requireMap(t, allow, "spec")
	requireNonEmptySlice(t, allowSpec, "ingress")
	requireNonEmptySlice(t, allowSpec, "egress")
}

func TestContainerPackaging(t *testing.T) {
	dockerfile := read(t, "deploy/inference-lab/Dockerfile")
	var instructions []string
	for _, line := range strings.Split(dockerfile, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			instructions = append(instructions, strings.ToUpper(line))
		}
	}
	effective := strings.Join(instructions, "\n")
	for _, token := range []string{"FROM ", " AS ", "COPY --from=", "USER ", "ENTRYPOINT"} {
		if !strings.Contains(effective, strings.ToUpper(token)) {
			t.Fatalf("Dockerfile must demonstrate %q", token)
		}
	}
	objects := parseYAML(t, read(t, "deploy/inference-lab/compose.yaml"))
	if len(objects) != 1 {
		t.Fatal("Compose file must contain one document")
	}
	services := requireMap(t, objects[0], "services")
	for _, name := range []string{"router", "model-a", "model-b"} {
		service := requireMap(t, services, name)
		requireMap(t, service, "healthcheck")
		if service["read_only"] != true {
			t.Errorf("Compose service %s must use a read-only root filesystem", name)
		}
		if service["privileged"] == true {
			t.Errorf("Compose service %s must not be privileged", name)
		}
		if !sliceContains(valueAt(service, "cap_drop"), "ALL") {
			t.Errorf("Compose service %s must drop all Linux capabilities", name)
		}
		requireMap(t, service, "deploy", "resources", "limits")
	}
}

func TestHelm(t *testing.T) {
	objects := render(t, "helm", "template", "inference-lab", "deploy/inference-lab/chart",
		"-f", "deploy/inference-lab/chart/values-learning.yaml")
	for _, required := range [][2]string{
		{"Deployment", "inference-lab-router"},
		{"Service", "inference-lab-router"},
		{"ConfigMap", "inference-lab-router"},
		{"ServiceAccount", "inference-lab-router"},
	} {
		requireObject(t, objects, required[0], required[1])
	}
	router := requireObject(t, objects, "Deployment", "inference-lab-router")
	requireProductionDeployment(t, router)
	for _, model := range []string{"model-a", "model-b"} {
		name := "inference-lab-" + model
		deployment := requireObject(t, objects, "Deployment", name)
		requireProductionDeployment(t, deployment)
		requireSelectorContract(t, deployment)
		requireServiceContract(t, requireObject(t, objects, "Service", name))
	}
	pdb := requireObject(t, objects, "PodDisruptionBudget", "inference-lab-router")
	requireOneOf(t, requireMap(t, pdb, "spec"), "minAvailable", "maxUnavailable")
	hpa := requireObject(t, objects, "HorizontalPodAutoscaler", "inference-lab-router")
	requireNonEmptySlice(t, requireMap(t, hpa, "spec"), "metrics")
	network := requireObject(t, objects, "NetworkPolicy", "inference-lab-router")
	requireNonEmptySlice(t, requireMap(t, network, "spec"), "policyTypes")
	var hookFound bool
	for _, candidate := range objects {
		if stringAt(candidate, "kind") != "Pod" {
			continue
		}
		annotations, _ := valueAt(candidate, "metadata", "annotations").(map[string]any)
		if annotations["helm.sh/hook"] == "test" || annotations["helm.sh/hook"] == "test-success" {
			hookFound = true
		}
	}
	if !hookFound {
		t.Fatal("Chart must render a Helm test Pod")
	}
}

func requireProductionDeployment(t *testing.T, deployment object) {
	t.Helper()
	container := requireFirstMap(t, deployment, "spec", "template", "spec", "containers")
	requireMap(t, requireMap(t, container, "readinessProbe"), "httpGet")
	requireMap(t, requireMap(t, container, "livenessProbe"), "httpGet")
	requireMap(t, container, "resources", "requests")
	requireMap(t, container, "resources", "limits")
	security := requireMap(t, container, "securityContext")
	for key, want := range map[string]bool{
		"allowPrivilegeEscalation": false,
		"readOnlyRootFilesystem":   true,
		"runAsNonRoot":             true,
	} {
		if security[key] != want {
			t.Errorf("container securityContext.%s = %v, want %v", key, security[key], want)
		}
	}
	if !sliceContains(valueAt(security, "capabilities", "drop"), "ALL") {
		t.Error("container must drop all Linux capabilities")
	}
}

func requireSelectorContract(t *testing.T, deployment object) {
	t.Helper()
	selector := requireMap(t, deployment, "spec", "selector", "matchLabels")
	labels := requireMap(t, deployment, "spec", "template", "metadata", "labels")
	for key, value := range selector {
		if labels[key] != value {
			t.Errorf("Deployment selector %s=%v does not match Pod label %v", key, value, labels[key])
		}
	}
}

func requireServiceContract(t *testing.T, service object) {
	t.Helper()
	requireMap(t, service, "spec", "selector")
	ports := requireNonEmptySlice(t, requireMap(t, service, "spec"), "ports")
	for index, raw := range ports {
		port, ok := raw.(map[string]any)
		if !ok || strings.TrimSpace(stringAt(port, "name")) == "" {
			t.Errorf("Service port %d must be named", index)
		}
	}
}

func sliceContains(value any, want string) bool {
	items, ok := value.([]any)
	if !ok {
		return false
	}
	for _, item := range items {
		if fmt.Sprint(item) == want {
			return true
		}
	}
	return false
}

func TestIstio(t *testing.T) {
	objects := render(t, "kubectl", "kustomize", "deploy/inference-lab/istio")
	gateway := requireObject(t, objects, "Gateway", "inference-gateway")
	listeners := requireNonEmptySlice(t, requireMap(t, gateway, "spec"), "listeners")
	for _, raw := range listeners {
		listener, ok := raw.(map[string]any)
		if !ok || listener["protocol"] != "HTTPS" {
			t.Fatal("public Gateway listeners must use HTTPS")
		}
		requireMap(t, listener, "tls")
	}
	route := requireObject(t, objects, "HTTPRoute", "inference-router")
	routeSpec := requireMap(t, route, "spec")
	requireNonEmptySlice(t, routeSpec, "parentRefs")
	for index, raw := range requireNonEmptySlice(t, routeSpec, "rules") {
		rule, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("HTTPRoute rule %d must be a map", index)
		}
		backends := requireNonEmptySlice(t, rule, "backendRefs")
		for _, backendRaw := range backends {
			backend := backendRaw.(map[string]any)
			if strings.TrimSpace(stringAt(backend, "name")) == "" {
				t.Fatal("every HTTPRoute backendRef must name a backend")
			}
		}
	}
	routeJSON, _ := json.Marshal(routeSpec)
	routeText := string(routeJSON)
	for _, forbidden := range []string{"/metrics", "/debug", "/readyz"} {
		if strings.Contains(routeText, forbidden) {
			t.Fatalf("public HTTPRoute exposes private path %s", forbidden)
		}
	}
	if !strings.Contains(routeText, "/v1") || !strings.Contains(routeText, "/livez") {
		t.Fatal("public HTTPRoute must explicitly match inference and liveness paths")
	}
	peer := requireObject(t, objects, "PeerAuthentication", "inference-lab-strict")
	if got := stringAt(peer, "spec", "mtls", "mode"); got != "STRICT" {
		t.Fatalf("PeerAuthentication mTLS mode = %q, want STRICT", got)
	}
	policy := requireObject(t, objects, "AuthorizationPolicy", "inference-router")
	policyRules := requireNonEmptySlice(t, requireMap(t, policy, "spec"), "rules")
	policyText, _ := json.Marshal(policyRules)
	if !strings.Contains(string(policyText), "principals") || !strings.Contains(string(policyText), "operations") {
		t.Fatal("AuthorizationPolicy must constrain both source principals and allowed operations")
	}
	var mutualTLS bool
	for _, candidate := range objects {
		if stringAt(candidate, "kind") == "DestinationRule" &&
			stringAt(candidate, "spec", "trafficPolicy", "tls", "mode") == "ISTIO_MUTUAL" {
			mutualTLS = true
		}
	}
	if !mutualTLS {
		t.Fatal("model traffic needs an explicit ISTIO_MUTUAL DestinationRule")
	}
}

func TestOperator(t *testing.T) {
	objects := render(t, "kubectl", "kustomize", "deploy/inference-lab/observability")
	monitor := requireObject(t, objects, "ServiceMonitor", "inference-router")
	monitorSpec := requireMap(t, monitor, "spec")
	requireMap(t, monitorSpec, "selector")
	endpoint := requireFirstMap(t, monitor, "spec", "endpoints")
	if endpoint["port"] != "http" || endpoint["path"] != "/metrics" {
		t.Fatalf("ServiceMonitor endpoint = %+v, want named http port and /metrics", endpoint)
	}
	rules := requireObject(t, objects, "PrometheusRule", "inference-router")
	requireRuleGroups(t, requireMap(t, rules, "spec"))
	dashboard := requireObject(t, objects, "ConfigMap", "inference-router-dashboard")
	data := requireMap(t, dashboard, "data")
	if len(data) == 0 {
		t.Fatal("Grafana dashboard ConfigMap data must not be empty")
	}
}

func TestLocalObservability(t *testing.T) {
	compose := parseYAML(t, read(t, "deploy/inference-lab/observability/compose.yaml"))
	services := requireMap(t, compose[0], "services")
	requireMap(t, services, "prometheus")
	requireMap(t, services, "grafana")

	prometheus := parseYAML(t, read(t, "deploy/inference-lab/observability/prometheus/prometheus.yml"))[0]
	requireNonEmptySlice(t, prometheus, "scrape_configs")
	provisioning := parseYAML(t, read(t,
		"deploy/inference-lab/observability/grafana/provisioning/datasources/prometheus.yaml"))[0]
	requireNonEmptySlice(t, provisioning, "datasources")

	var dashboard object
	if err := json.Unmarshal([]byte(read(t,
		"deploy/inference-lab/observability/grafana/dashboards/inference-router.json")), &dashboard); err != nil {
		t.Fatalf("parse dashboard JSON: %v", err)
	}
	panels := requireNonEmptySlice(t, dashboard, "panels")
	for index, raw := range panels {
		panel, ok := raw.(map[string]any)
		if !ok || strings.TrimSpace(stringAt(panel, "title")) == "" ||
			strings.TrimSpace(stringAt(panel, "type")) == "" {
			t.Errorf("dashboard panel %d needs a title and type", index)
			continue
		}
		requireNonEmptySlice(t, panel, "targets")
	}
}

func TestPrometheusRules(t *testing.T) {
	rules := parseYAML(t, read(t, "deploy/inference-lab/observability/prometheus/rules.yml"))
	requireRuleGroups(t, rules[0])
}

func TestGAIE(t *testing.T) {
	objects := render(t, "kubectl", "kustomize", "deploy/inference-lab/gaie")
	pool := requireObject(t, objects, "InferencePool", "inference-lab")
	if stringAt(pool, "apiVersion") != "inference.networking.k8s.io/v1" {
		t.Fatalf("InferencePool apiVersion = %q, want stable v1", stringAt(pool, "apiVersion"))
	}
	spec := requireMap(t, pool, "spec")
	requireMap(t, spec, "selector", "matchLabels")
	for index, raw := range requireNonEmptySlice(t, spec, "targetPorts") {
		port, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("targetPorts[%d] must be a Port object", index)
		}
		requirePositiveNumber(t, port, "number")
	}
	picker := requireMap(t, spec, "endpointPickerRef")
	if strings.TrimSpace(stringAt(picker, "name")) == "" {
		t.Fatal("endpointPickerRef.name is required for this llm-d exercise")
	}
	requirePositiveNumber(t, requireMap(t, picker, "port"), "number")
	if mode := stringAt(picker, "failureMode"); mode != "FailOpen" && mode != "FailClose" {
		t.Fatalf("endpointPickerRef.failureMode = %q, want an explicit supported mode", mode)
	}
	route := requireObject(t, objects, "HTTPRoute", "inference-pool")
	routeJSON, _ := json.Marshal(requireMap(t, route, "spec"))
	for _, token := range []string{"inference.networking.k8s.io", "InferencePool", "inference-lab"} {
		if !strings.Contains(string(routeJSON), token) {
			t.Fatalf("InferencePool HTTPRoute backend is missing %q", token)
		}
	}
}

func TestRuntimeConfig(t *testing.T) {
	objects := parseYAML(t, read(t, "configs/endpoint-picker-learning.yaml"))
	config := requireObject(t, objects, "EndpointPickerConfig", "inference-lab")
	spec := requireMap(t, config, "spec")
	requireMap(t, spec, "requestHandling")
	dataLayer := requireMap(t, spec, "dataLayer")
	requireNonEmptySlice(t, dataLayer, "sources")
	flow := requireMap(t, spec, "flowControl")
	requirePositiveNumber(t, flow, "maxInFlight")
	requirePositiveNumber(t, flow, "maxQueued")
	requireNonEmptySlice(t, spec, "schedulingProfiles")
	requireMap(t, spec, "failurePolicy")
}

func TestRobotContracts(t *testing.T) {
	resource := read(t, "test/acceptance/inference/resources/inference.resource")
	if strings.Contains(resource, "TODO(exercise)") || strings.Contains(resource, "\n    Fail") {
		t.Error("shared Robot resource still contains a placeholder implementation")
	}
	for _, operation := range []string{"Create Session", "POST On Session", "GET On Session"} {
		if !strings.Contains(resource, operation) {
			t.Errorf("shared Robot resource must implement public HTTP operation %q", operation)
		}
	}
	for _, relative := range []string{
		"test/acceptance/inference/00_smoke.robot",
		"test/acceptance/inference/01_routing.robot",
		"test/acceptance/inference/02_negative.robot",
		"test/acceptance/inference/03_production.robot",
	} {
		content := read(t, relative)
		if strings.Contains(content, "TODO(exercise)") || strings.Contains(content, "\n    Fail") {
			t.Errorf("%s still contains a placeholder failure", relative)
			continue
		}
		cases := robotCases(content)
		if len(cases) == 0 {
			t.Errorf("%s contains no test cases", relative)
			continue
		}
		for name, body := range cases {
			if len(body) < 2 {
				t.Errorf("%s: %q needs both an interaction and an observable assertion", relative, name)
				continue
			}
			joined := strings.Join(body, "\n")
			hasInteraction := containsOne(joined, "Route Chat", "GET ", "POST ", "Read Last Decision")
			hasAssertion := containsOne(joined, "Should", "Wait Until", "Expect Error")
			if !hasInteraction || !hasAssertion {
				t.Errorf("%s: %q must exercise public HTTP behavior and assert an observation", relative, name)
			}
		}
	}
}

func robotCases(content string) map[string][]string {
	parts := strings.SplitN(content, "*** Test Cases ***", 2)
	if len(parts) != 2 {
		return nil
	}
	cases := map[string][]string{}
	var current string
	for _, line := range strings.Split(parts[1], "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if line == trimmed {
			current = trimmed
			cases[current] = nil
			continue
		}
		if current != "" && !strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "...") {
			cases[current] = append(cases[current], trimmed)
		}
	}
	return cases
}

func containsOne(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func render(t *testing.T, program string, args ...string) []object {
	t.Helper()
	command := exec.Command(program, args...)
	command.Dir = root(t)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", program, err, output)
	}
	return parseYAML(t, string(output))
}

func parseYAML(t *testing.T, content string) []object {
	t.Helper()
	decoder := yaml.NewDecoder(strings.NewReader(content))
	var objects []object
	for {
		var item object
		err := decoder.Decode(&item)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("parse YAML: %v", err)
		}
		if len(item) > 0 {
			objects = append(objects, item)
		}
	}
	return objects
}

func requireObject(t *testing.T, objects []object, kind, name string) object {
	t.Helper()
	for _, item := range objects {
		if stringAt(item, "kind") == kind && stringAt(item, "metadata", "name") == name {
			return item
		}
	}
	t.Fatalf("rendered resources do not contain %s/%s", kind, name)
	return nil
}

func valueAt(value any, path ...string) any {
	current := value
	for _, key := range path {
		mapping, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = mapping[key]
	}
	return current
}

func stringAt(value any, path ...string) string {
	result, _ := valueAt(value, path...).(string)
	return result
}

func requireMap(t *testing.T, value any, path ...string) object {
	t.Helper()
	result, ok := valueAt(value, path...).(map[string]any)
	if !ok || len(result) == 0 {
		t.Fatalf("%s must be a non-empty map", strings.Join(path, "."))
	}
	return result
}

func requireNonEmptySlice(t *testing.T, value any, key string) []any {
	t.Helper()
	result, ok := valueAt(value, key).([]any)
	if !ok || len(result) == 0 {
		t.Fatalf("%s must be a non-empty list", key)
	}
	return result
}

func requireFirstMap(t *testing.T, value any, path ...string) object {
	t.Helper()
	items, ok := valueAt(value, path...).([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("%s must contain at least one item", strings.Join(path, "."))
	}
	item, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("%s[0] must be a map", strings.Join(path, "."))
	}
	return item
}

func requireOneOf(t *testing.T, value object, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if _, ok := value[key]; ok {
			return
		}
	}
	t.Fatalf("one of %v is required", keys)
}

func requireNumberAtLeast(t *testing.T, value any, minimum float64, path ...string) {
	t.Helper()
	raw := valueAt(value, path...)
	number, ok := asNumber(raw)
	if !ok || number < minimum {
		t.Fatalf("%s = %v, want >= %v", strings.Join(path, "."), raw, minimum)
	}
}

func requirePositiveNumber(t *testing.T, value object, key string) {
	t.Helper()
	number, ok := asNumber(value[key])
	if !ok || number <= 0 {
		t.Fatalf("%s = %v, want a positive number", key, value[key])
	}
}

func asNumber(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case float64:
		return typed, true
	case string:
		number, err := strconv.ParseFloat(typed, 64)
		return number, err == nil
	default:
		return 0, false
	}
}

func requireRuleGroups(t *testing.T, value object) {
	t.Helper()
	groups := requireNonEmptySlice(t, value, "groups")
	var recordingRule, alertRule bool
	for index, raw := range groups {
		group, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("groups[%d] must be a map", index)
		}
		for ruleIndex, rawRule := range requireNonEmptySlice(t, group, "rules") {
			rule, ok := rawRule.(map[string]any)
			if !ok {
				t.Fatalf("groups[%d].rules[%d] must be a map", index, ruleIndex)
			}
			if _, ok := rule["record"]; ok {
				recordingRule = true
			}
			if strings.TrimSpace(fmt.Sprint(rule["expr"])) == "" ||
				fmt.Sprint(rule["expr"]) == "<nil>" {
				t.Fatalf("rule %d/%d requires a non-empty expression", index, ruleIndex)
			}
			if _, ok := rule["alert"]; ok {
				alertRule = true
				pending, err := time.ParseDuration(stringAt(rule, "for"))
				if err != nil || pending <= 0 {
					t.Fatalf("alert %v must use a positive pending period", rule["alert"])
				}
				requireMap(t, rule, "labels")
				requireMap(t, rule, "annotations")
			}
		}
	}
	if !recordingRule || !alertRule {
		t.Fatalf("rules need at least one recording rule and one alert; recording=%v alert=%v",
			recordingRule, alertRule)
	}
}

func read(t *testing.T, relative string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root(t), filepath.FromSlash(relative)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func root(t *testing.T) string {
	t.Helper()
	working, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.Abs(filepath.Join(working, "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}
