// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package shared

import "errors"

// ValidateUserID matches the public API's toApiPathOrQueryParamNonNegativeInt:
// user IDs are integers >= 0. Cobra's integer flags reject fractional input.
// See codesphere-monorepo/packages/public-api-service/common/src/ts/server/apiTypeConverter.ts.
func ValidateUserID(userID int) error {
	if userID < 0 {
		return errors.New("user ID must be non-negative")
	}
	return nil
}
