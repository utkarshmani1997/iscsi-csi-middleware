package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Secrets provides optional iscsi security credentials (CHAP settings)
// +k8s:openapi-gen=true
type Secrets struct {
	// SecretsType is the type of Secrets being utilized (currently we only impleemnent "chap"
	SecretsType string `json:"secretsType,omitempty"`
	// UserName is the configured iscsi user login
	UserName string `json:"userName"`
	// Password is the configured iscsi password
	Password string `json:"password"`
	// UserNameIn provides a specific input login for directional CHAP configurations
	UserNameIn string `json:"userNameIn,omitempty"`
	// PasswordIn provides a specific input password for directional CHAP configurations
	PasswordIn string `json:"passwordIn,omitempty"`
}

// ISCSIConnectionSpec defines the desired state of ISCSIConnection
// +k8s:openapi-gen=true
type ISCSIConnectionSpec struct {
	VolumeName         string   `json:"volume_name"`
	TargetIqn          string   `json:"target_iqn"`
	TargetPortals      []string `json:"target_portals"`
	Port               string   `json:"port"`
	Lun                int32    `json:"lun"`
	AuthType           string   `json:"auth_type"`
	NodeName           string   `json:"node_name"`
	DiscoverySecrets   Secrets  `json:"discovery_secrets"`
	SessionSecrets     Secrets  `json:"session_secrets"`
	Interface          string   `json:"interface"`
	Multipath          bool     `json:"multipath"`
	RetryCount         int32    `json:"retry_count"`
	CheckInterval      int32    `json:"check_interval"`
	ReplacementTimeout string   `json:"replacement_timeout"`
	IsFormatted        bool     `json:"isFormated"`
}

// ISCSIConnectionPhase represents the current phase of JivaVolume.
type ISCSIConnectionPhase string

const (
	// ISCSIConnectionPhasePending indicates that the iscsi connection is yet to
	// establish.
	ISCSIConnectionPhasePending ISCSIConnectionPhase = "Pending"

	// ISCSIConnectionPhaseLoginFailed indicates that iscsi connection is failed to
	// login with target.
	ISCSIConnectionPhaseLoginFailed ISCSIConnectionPhase = "LoginFailed"

	// ISCSIConnectionPhaseLoginSuccess indicates that iscsi connection is
	// established with the target and login successful
	ISCSIConnectionPhaseLoginSuccess ISCSIConnectionPhase = "LoginSuccessful"

	// ISCSIConnectionPhaseLogoutFailed indicates the the logout is failed with
	// the iscsi target
	ISCSIConnectionPhaseLogoutFailed ISCSIConnectionPhase = "LogoutFailed"

	// ISCSIConnectionPhaseLogoutSuccess indicates that iscsi connection is
	// established with the target and login successful
	ISCSIConnectionPhaseLogoutSuccess ISCSIConnectionPhase = "LogoutSuccessful"

	// ISCSIConnectionPhaseLogoutStart indicates that iscsi logout is requested
	ISCSIConnectionPhaseLogoutStart ISCSIConnectionPhase = "LogoutStart"
)

// ISCSIConnectionStatus defines the observed state of ISCSIConnection
// +k8s:openapi-gen=true
type ISCSIConnectionStatus struct {
	Status string               `json:"status"`
	Phase  ISCSIConnectionPhase `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ISCSIConnection is the Schema for the iscsiconnections API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ISCSIConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ISCSIConnectionSpec   `json:"spec,omitempty"`
	Status ISCSIConnectionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ISCSIConnectionList contains a list of ISCSIConnection
type ISCSIConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ISCSIConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ISCSIConnection{}, &ISCSIConnectionList{})
}
