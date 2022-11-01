//+groupName="install.terraform.io"
//+versionName="v1"

// Package types defines structures for installer configuration and
// management.
package types

// Architecture is the instruction set architecture for the machines in a pool.
// +kubebuilder:validation:Enum="";amd64
type Architecture string

const (
	// ArchitectureAMD64 indicates AMD64 (x86_64).
	ArchitectureAMD64 = "amd64"
	// ArchitectureS390X indicates s390x (IBM System Z).
	ArchitectureS390X = "s390x"
	// ArchitecturePPC64LE indicates ppc64 little endian (Power PC)
	ArchitecturePPC64LE = "ppc64le"
	// ArchitectureARM64 indicates arm (aarch64) systems
	ArchitectureARM64 = "arm64"
)
