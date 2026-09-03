package proto_test

import (
	"testing"

	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
	agentv1alpha1 "github.com/kastellan/kastellan/api/proto/kastellan/agent/v1alpha1"
)

func TestContainerInfoToProto(t *testing.T) {
	container := &messages.ContainerInfo{
		Name:  "test-container",
		ID:    "container-id-123",
		State: "running",
	}

	proto := container.ToProto()

	if proto.GetName() != "test-container" {
		t.Errorf("expected name 'test-container', got '%s'", proto.GetName())
	}
	if proto.GetId() != "container-id-123" {
		t.Errorf("expected id 'container-id-123', got '%s'", proto.GetId())
	}
	if proto.GetState() != "running" {
		t.Errorf("expected state 'running', got '%s'", proto.GetState())
	}
}

func TestContainerInfoFromProto(t *testing.T) {
	proto := &agentv1alpha1.ContainerInfo{
		Name:  "test-container",
		Id:    "container-id-123",
		State: "running",
	}

	container := messages.ContainerInfoFromProto(proto)

	if container.Name != "test-container" {
		t.Errorf("expected name 'test-container', got '%s'", container.Name)
	}
	if container.ID != "container-id-123" {
		t.Errorf("expected id 'container-id-123', got '%s'", container.ID)
	}
	if container.State != "running" {
		t.Errorf("expected state 'running', got '%s'", container.State)
	}
}

func TestContainerInfoRoundTrip(t *testing.T) {
	original := &messages.ContainerInfo{
		Name:  "test-container",
		ID:    "container-id-123",
		State: "running",
	}

	proto := original.ToProto()
	back := messages.ContainerInfoFromProto(proto)

	if original.Name != back.Name {
		t.Errorf("name mismatch: original=%s, back=%s", original.Name, back.Name)
	}
	if original.ID != back.ID {
		t.Errorf("ID mismatch: original=%s, back=%s", original.ID, back.ID)
	}
	if original.State != back.State {
		t.Errorf("state mismatch: original=%s, back=%s", original.State, back.State)
	}
}

func TestWorkloadResultToProto(t *testing.T) {
	original := &messages.WorkloadResult{
		UID:        "workload-uid-123",
		Namespace:  "default",
		Name:       "test-workload",
		Generation: 1,
		Phase:      "Running",
		ManifestDigest: "sha256:abc123",
	}
	original.Runtime.Containers = []messages.ContainerInfo{
		{Name: "container1", ID: "container1-id", State: "running"},
	}

	proto := original.ToProto()

	if proto.GetIdentity().GetUid() != "workload-uid-123" {
		t.Errorf("expected uid 'workload-uid-123', got '%s'", proto.GetIdentity().GetUid())
	}
	if proto.GetPhase() != "Running" {
		t.Errorf("expected phase 'Running', got '%s'", proto.GetPhase())
	}
	if proto.GetManifestDigest() != "sha256:abc123" {
		t.Errorf("expected manifest digest 'sha256:abc123', got '%s'", proto.GetManifestDigest())
	}
}

func TestPodmanPlayToProto(t *testing.T) {
	original := &messages.PodmanPlay{
		UID:        "play-uid-123",
		Namespace:  "default",
		Name:       "test-play",
		Generation: 1,
		Manifest:   "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test",
	}

	proto := original.ToProto()

	if proto.GetIdentity().GetUid() != "play-uid-123" {
		t.Errorf("expected uid 'play-uid-123', got '%s'", proto.GetIdentity().GetUid())
	}
	if proto.GetManifest() != "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test" {
		t.Errorf("manifest mismatch")
	}
}
