package server

import (
	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
)

// Test types for testing purposes.
// In production, use the actual protobuf-generated types.

// TestAgentMessage is a test agent message type.
type TestAgentMessage struct {
	Type messages.MessageType
}

// TestOperatorMessage is a test operator message type.
type TestOperatorMessage struct {
	Type messages.MessageType
}

// TestResourceIdentity is a test resource identity type.
type TestResourceIdentity struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
	UID        string
	Generation int64
}

// TestWorkloadResult is a test workload result type.
type TestWorkloadResult struct {
	UID            string
	Namespace      string
	Name           string
	Generation     int64
	Phase          string
	Reason         string
	Message        string
	ManifestDigest string
	Runtime        struct {
		PodID      string
		Containers []messages.ContainerInfo
	}
}

// TestWorkloadStatus is a test workload status type.
type TestWorkloadStatus struct {
	UID        string
	Namespace  string
	Name       string
	Generation int64
	Phase      string
	Runtime    struct {
		PodID      string
		Containers []messages.ContainerInfo
	}
}
