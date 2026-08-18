// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver

// itemsResult wraps a list result for an MCP tool response, making sure
// "items" is serialized as an empty array instead of null when the list
// is empty or the API returned a nil slice.
func itemsResult[T any](items []T) map[string]any {
	if items == nil {
		items = []T{}
	}
	return map[string]any{"items": items}
}
