// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

/*
Codesphere Public API

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi_client

import (
	"encoding/json"
	"fmt"
	"gopkg.in/validator.v2"
)

// WorkspacesReplicaLogs200Response - SSE stream with two event types: \"data\" and \"problem\". Both event data contain JSON objects in the form described by their schemas. Possible problem statuses and reasons:400: Workspace is not running, path or request body variable does not match schema. 401: Authorization information is missing or invalid. 404: Workspace is not found.
type WorkspacesReplicaLogs200Response struct {
	Problem                          *Problem
	WorkspacesReplicaLogsGetResponse *WorkspacesReplicaLogsGetResponse
}

// ProblemAsWorkspacesReplicaLogs200Response is a convenience function that returns Problem wrapped in WorkspacesReplicaLogs200Response
func ProblemAsWorkspacesReplicaLogs200Response(v *Problem) WorkspacesReplicaLogs200Response {
	return WorkspacesReplicaLogs200Response{
		Problem: v,
	}
}

// WorkspacesReplicaLogsGetResponseAsWorkspacesReplicaLogs200Response is a convenience function that returns WorkspacesReplicaLogsGetResponse wrapped in WorkspacesReplicaLogs200Response
func WorkspacesReplicaLogsGetResponseAsWorkspacesReplicaLogs200Response(v *WorkspacesReplicaLogsGetResponse) WorkspacesReplicaLogs200Response {
	return WorkspacesReplicaLogs200Response{
		WorkspacesReplicaLogsGetResponse: v,
	}
}

// Unmarshal JSON data into one of the pointers in the struct
func (dst *WorkspacesReplicaLogs200Response) UnmarshalJSON(data []byte) error {
	var err error
	match := 0
	// try to unmarshal data into Problem
	err = newStrictDecoder(data).Decode(&dst.Problem)
	if err == nil {
		jsonProblem, _ := json.Marshal(dst.Problem)
		if string(jsonProblem) == "{}" { // empty struct
			dst.Problem = nil
		} else {
			if err = validator.Validate(dst.Problem); err != nil {
				dst.Problem = nil
			} else {
				match++
			}
		}
	} else {
		dst.Problem = nil
	}

	// try to unmarshal data into WorkspacesReplicaLogsGetResponse
	err = newStrictDecoder(data).Decode(&dst.WorkspacesReplicaLogsGetResponse)
	if err == nil {
		jsonWorkspacesReplicaLogsGetResponse, _ := json.Marshal(dst.WorkspacesReplicaLogsGetResponse)
		if string(jsonWorkspacesReplicaLogsGetResponse) == "{}" { // empty struct
			dst.WorkspacesReplicaLogsGetResponse = nil
		} else {
			if err = validator.Validate(dst.WorkspacesReplicaLogsGetResponse); err != nil {
				dst.WorkspacesReplicaLogsGetResponse = nil
			} else {
				match++
			}
		}
	} else {
		dst.WorkspacesReplicaLogsGetResponse = nil
	}

	if match > 1 { // more than 1 match
		// reset to nil
		dst.Problem = nil
		dst.WorkspacesReplicaLogsGetResponse = nil

		return fmt.Errorf("data matches more than one schema in oneOf(WorkspacesReplicaLogs200Response)")
	} else if match == 1 {
		return nil // exactly one match
	} else { // no match
		return fmt.Errorf("data failed to match schemas in oneOf(WorkspacesReplicaLogs200Response)")
	}
}

// Marshal data from the first non-nil pointers in the struct to JSON
func (src WorkspacesReplicaLogs200Response) MarshalJSON() ([]byte, error) {
	if src.Problem != nil {
		return json.Marshal(&src.Problem)
	}

	if src.WorkspacesReplicaLogsGetResponse != nil {
		return json.Marshal(&src.WorkspacesReplicaLogsGetResponse)
	}

	return nil, nil // no data in oneOf schemas
}

// Get the actual instance
func (obj *WorkspacesReplicaLogs200Response) GetActualInstance() interface{} {
	if obj == nil {
		return nil
	}
	if obj.Problem != nil {
		return obj.Problem
	}

	if obj.WorkspacesReplicaLogsGetResponse != nil {
		return obj.WorkspacesReplicaLogsGetResponse
	}

	// all schemas are nil
	return nil
}

// Get the actual instance value
func (obj WorkspacesReplicaLogs200Response) GetActualInstanceValue() interface{} {
	if obj.Problem != nil {
		return *obj.Problem
	}

	if obj.WorkspacesReplicaLogsGetResponse != nil {
		return *obj.WorkspacesReplicaLogsGetResponse
	}

	// all schemas are nil
	return nil
}

type NullableWorkspacesReplicaLogs200Response struct {
	value *WorkspacesReplicaLogs200Response
	isSet bool
}

func (v NullableWorkspacesReplicaLogs200Response) Get() *WorkspacesReplicaLogs200Response {
	return v.value
}

func (v *NullableWorkspacesReplicaLogs200Response) Set(val *WorkspacesReplicaLogs200Response) {
	v.value = val
	v.isSet = true
}

func (v NullableWorkspacesReplicaLogs200Response) IsSet() bool {
	return v.isSet
}

func (v *NullableWorkspacesReplicaLogs200Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableWorkspacesReplicaLogs200Response(val *WorkspacesReplicaLogs200Response) *NullableWorkspacesReplicaLogs200Response {
	return &NullableWorkspacesReplicaLogs200Response{value: val, isSet: true}
}

func (v NullableWorkspacesReplicaLogs200Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableWorkspacesReplicaLogs200Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
