// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

/*
Codesphere Public API

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi_client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// checks if the WorkspacesCreateWorkspaceRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &WorkspacesCreateWorkspaceRequest{}

// WorkspacesCreateWorkspaceRequest struct for WorkspacesCreateWorkspaceRequest
type WorkspacesCreateWorkspaceRequest struct {
	TeamId            int     `json:"teamId"`
	Name              string  `json:"name"`
	PlanId            int     `json:"planId"`
	IsPrivateRepo     bool    `json:"isPrivateRepo"`
	Replicas          int     `json:"replicas"`
	GitUrl            *string `json:"gitUrl,omitempty"`
	InitialBranch     *string `json:"initialBranch,omitempty"`
	SourceWorkspaceId *int    `json:"sourceWorkspaceId,omitempty"`
	WelcomeMessage    *string `json:"welcomeMessage,omitempty"`
	VpnConfig         *string `json:"vpnConfig,omitempty"`
}

type _WorkspacesCreateWorkspaceRequest WorkspacesCreateWorkspaceRequest

// NewWorkspacesCreateWorkspaceRequest instantiates a new WorkspacesCreateWorkspaceRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewWorkspacesCreateWorkspaceRequest(teamId int, name string, planId int, isPrivateRepo bool, replicas int) *WorkspacesCreateWorkspaceRequest {
	this := WorkspacesCreateWorkspaceRequest{}
	this.TeamId = teamId
	this.Name = name
	this.PlanId = planId
	this.IsPrivateRepo = isPrivateRepo
	this.Replicas = replicas
	return &this
}

// NewWorkspacesCreateWorkspaceRequestWithDefaults instantiates a new WorkspacesCreateWorkspaceRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewWorkspacesCreateWorkspaceRequestWithDefaults() *WorkspacesCreateWorkspaceRequest {
	this := WorkspacesCreateWorkspaceRequest{}
	return &this
}

// GetTeamId returns the TeamId field value
func (o *WorkspacesCreateWorkspaceRequest) GetTeamId() int {
	if o == nil {
		var ret int
		return ret
	}

	return o.TeamId
}

// GetTeamIdOk returns a tuple with the TeamId field value
// and a boolean to check if the value has been set.
func (o *WorkspacesCreateWorkspaceRequest) GetTeamIdOk() (*int, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TeamId, true
}

// SetTeamId sets field value
func (o *WorkspacesCreateWorkspaceRequest) SetTeamId(v int) {
	o.TeamId = v
}

// GetName returns the Name field value
func (o *WorkspacesCreateWorkspaceRequest) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *WorkspacesCreateWorkspaceRequest) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *WorkspacesCreateWorkspaceRequest) SetName(v string) {
	o.Name = v
}

// GetPlanId returns the PlanId field value
func (o *WorkspacesCreateWorkspaceRequest) GetPlanId() int {
	if o == nil {
		var ret int
		return ret
	}

	return o.PlanId
}

// GetPlanIdOk returns a tuple with the PlanId field value
// and a boolean to check if the value has been set.
func (o *WorkspacesCreateWorkspaceRequest) GetPlanIdOk() (*int, bool) {
	if o == nil {
		return nil, false
	}
	return &o.PlanId, true
}

// SetPlanId sets field value
func (o *WorkspacesCreateWorkspaceRequest) SetPlanId(v int) {
	o.PlanId = v
}

// GetIsPrivateRepo returns the IsPrivateRepo field value
func (o *WorkspacesCreateWorkspaceRequest) GetIsPrivateRepo() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.IsPrivateRepo
}

// GetIsPrivateRepoOk returns a tuple with the IsPrivateRepo field value
// and a boolean to check if the value has been set.
func (o *WorkspacesCreateWorkspaceRequest) GetIsPrivateRepoOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.IsPrivateRepo, true
}

// SetIsPrivateRepo sets field value
func (o *WorkspacesCreateWorkspaceRequest) SetIsPrivateRepo(v bool) {
	o.IsPrivateRepo = v
}

// GetReplicas returns the Replicas field value
func (o *WorkspacesCreateWorkspaceRequest) GetReplicas() int {
	if o == nil {
		var ret int
		return ret
	}

	return o.Replicas
}

// GetReplicasOk returns a tuple with the Replicas field value
// and a boolean to check if the value has been set.
func (o *WorkspacesCreateWorkspaceRequest) GetReplicasOk() (*int, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Replicas, true
}

// SetReplicas sets field value
func (o *WorkspacesCreateWorkspaceRequest) SetReplicas(v int) {
	o.Replicas = v
}

// GetGitUrl returns the GitUrl field value if set, zero value otherwise.
func (o *WorkspacesCreateWorkspaceRequest) GetGitUrl() string {
	if o == nil || IsNil(o.GitUrl) {
		var ret string
		return ret
	}
	return *o.GitUrl
}

// GetGitUrlOk returns a tuple with the GitUrl field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WorkspacesCreateWorkspaceRequest) GetGitUrlOk() (*string, bool) {
	if o == nil || IsNil(o.GitUrl) {
		return nil, false
	}
	return o.GitUrl, true
}

// HasGitUrl returns a boolean if a field has been set.
func (o *WorkspacesCreateWorkspaceRequest) HasGitUrl() bool {
	if o != nil && !IsNil(o.GitUrl) {
		return true
	}

	return false
}

// SetGitUrl gets a reference to the given string and assigns it to the GitUrl field.
func (o *WorkspacesCreateWorkspaceRequest) SetGitUrl(v string) {
	o.GitUrl = &v
}

// GetInitialBranch returns the InitialBranch field value if set, zero value otherwise.
func (o *WorkspacesCreateWorkspaceRequest) GetInitialBranch() string {
	if o == nil || IsNil(o.InitialBranch) {
		var ret string
		return ret
	}
	return *o.InitialBranch
}

// GetInitialBranchOk returns a tuple with the InitialBranch field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WorkspacesCreateWorkspaceRequest) GetInitialBranchOk() (*string, bool) {
	if o == nil || IsNil(o.InitialBranch) {
		return nil, false
	}
	return o.InitialBranch, true
}

// HasInitialBranch returns a boolean if a field has been set.
func (o *WorkspacesCreateWorkspaceRequest) HasInitialBranch() bool {
	if o != nil && !IsNil(o.InitialBranch) {
		return true
	}

	return false
}

// SetInitialBranch gets a reference to the given string and assigns it to the InitialBranch field.
func (o *WorkspacesCreateWorkspaceRequest) SetInitialBranch(v string) {
	o.InitialBranch = &v
}

// GetSourceWorkspaceId returns the SourceWorkspaceId field value if set, zero value otherwise.
func (o *WorkspacesCreateWorkspaceRequest) GetSourceWorkspaceId() int {
	if o == nil || IsNil(o.SourceWorkspaceId) {
		var ret int
		return ret
	}
	return *o.SourceWorkspaceId
}

// GetSourceWorkspaceIdOk returns a tuple with the SourceWorkspaceId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WorkspacesCreateWorkspaceRequest) GetSourceWorkspaceIdOk() (*int, bool) {
	if o == nil || IsNil(o.SourceWorkspaceId) {
		return nil, false
	}
	return o.SourceWorkspaceId, true
}

// HasSourceWorkspaceId returns a boolean if a field has been set.
func (o *WorkspacesCreateWorkspaceRequest) HasSourceWorkspaceId() bool {
	if o != nil && !IsNil(o.SourceWorkspaceId) {
		return true
	}

	return false
}

// SetSourceWorkspaceId gets a reference to the given int and assigns it to the SourceWorkspaceId field.
func (o *WorkspacesCreateWorkspaceRequest) SetSourceWorkspaceId(v int) {
	o.SourceWorkspaceId = &v
}

// GetWelcomeMessage returns the WelcomeMessage field value if set, zero value otherwise.
func (o *WorkspacesCreateWorkspaceRequest) GetWelcomeMessage() string {
	if o == nil || IsNil(o.WelcomeMessage) {
		var ret string
		return ret
	}
	return *o.WelcomeMessage
}

// GetWelcomeMessageOk returns a tuple with the WelcomeMessage field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WorkspacesCreateWorkspaceRequest) GetWelcomeMessageOk() (*string, bool) {
	if o == nil || IsNil(o.WelcomeMessage) {
		return nil, false
	}
	return o.WelcomeMessage, true
}

// HasWelcomeMessage returns a boolean if a field has been set.
func (o *WorkspacesCreateWorkspaceRequest) HasWelcomeMessage() bool {
	if o != nil && !IsNil(o.WelcomeMessage) {
		return true
	}

	return false
}

// SetWelcomeMessage gets a reference to the given string and assigns it to the WelcomeMessage field.
func (o *WorkspacesCreateWorkspaceRequest) SetWelcomeMessage(v string) {
	o.WelcomeMessage = &v
}

// GetVpnConfig returns the VpnConfig field value if set, zero value otherwise.
func (o *WorkspacesCreateWorkspaceRequest) GetVpnConfig() string {
	if o == nil || IsNil(o.VpnConfig) {
		var ret string
		return ret
	}
	return *o.VpnConfig
}

// GetVpnConfigOk returns a tuple with the VpnConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WorkspacesCreateWorkspaceRequest) GetVpnConfigOk() (*string, bool) {
	if o == nil || IsNil(o.VpnConfig) {
		return nil, false
	}
	return o.VpnConfig, true
}

// HasVpnConfig returns a boolean if a field has been set.
func (o *WorkspacesCreateWorkspaceRequest) HasVpnConfig() bool {
	if o != nil && !IsNil(o.VpnConfig) {
		return true
	}

	return false
}

// SetVpnConfig gets a reference to the given string and assigns it to the VpnConfig field.
func (o *WorkspacesCreateWorkspaceRequest) SetVpnConfig(v string) {
	o.VpnConfig = &v
}

func (o WorkspacesCreateWorkspaceRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o WorkspacesCreateWorkspaceRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["teamId"] = o.TeamId
	toSerialize["name"] = o.Name
	toSerialize["planId"] = o.PlanId
	toSerialize["isPrivateRepo"] = o.IsPrivateRepo
	toSerialize["replicas"] = o.Replicas
	if !IsNil(o.GitUrl) {
		toSerialize["gitUrl"] = o.GitUrl
	}
	if !IsNil(o.InitialBranch) {
		toSerialize["initialBranch"] = o.InitialBranch
	}
	if !IsNil(o.SourceWorkspaceId) {
		toSerialize["sourceWorkspaceId"] = o.SourceWorkspaceId
	}
	if !IsNil(o.WelcomeMessage) {
		toSerialize["welcomeMessage"] = o.WelcomeMessage
	}
	if !IsNil(o.VpnConfig) {
		toSerialize["vpnConfig"] = o.VpnConfig
	}
	return toSerialize, nil
}

func (o *WorkspacesCreateWorkspaceRequest) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"teamId",
		"name",
		"planId",
		"isPrivateRepo",
		"replicas",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err
	}

	for _, requiredProperty := range requiredProperties {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varWorkspacesCreateWorkspaceRequest := _WorkspacesCreateWorkspaceRequest{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	//decoder.DisallowUnknownFields()
	err = decoder.Decode(&varWorkspacesCreateWorkspaceRequest)

	if err != nil {
		return err
	}

	*o = WorkspacesCreateWorkspaceRequest(varWorkspacesCreateWorkspaceRequest)

	return err
}

type NullableWorkspacesCreateWorkspaceRequest struct {
	value *WorkspacesCreateWorkspaceRequest
	isSet bool
}

func (v NullableWorkspacesCreateWorkspaceRequest) Get() *WorkspacesCreateWorkspaceRequest {
	return v.value
}

func (v *NullableWorkspacesCreateWorkspaceRequest) Set(val *WorkspacesCreateWorkspaceRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableWorkspacesCreateWorkspaceRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableWorkspacesCreateWorkspaceRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableWorkspacesCreateWorkspaceRequest(val *WorkspacesCreateWorkspaceRequest) *NullableWorkspacesCreateWorkspaceRequest {
	return &NullableWorkspacesCreateWorkspaceRequest{value: val, isSet: true}
}

func (v NullableWorkspacesCreateWorkspaceRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableWorkspacesCreateWorkspaceRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
