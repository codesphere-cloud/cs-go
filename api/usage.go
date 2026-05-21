package api

import (
	"time"

	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

func (c *Client) GetLandscapeServiceEvents(teamId int, resourceId string, beginDate, endDate time.Time, limit, offset int) (*openapi_client.UsageGetLandscapeServiceEvents200Response, error) {
	req := c.api.UsageAPI.UsageGetLandscapeServiceEvents(c.ctx, float32(teamId), resourceId).
		BeginDate(beginDate).
		EndDate(endDate)
	if limit > 0 {
		req = req.Limit(limit)
	}
	if offset > 0 {
		req = req.Offset(offset)
	}
	res, r, err := req.Execute()
	return res, errors.FormatAPIError(r, err)
}

func (c *Client) GetUsageSummaryLandscape(teamId int, beginDate, endDate time.Time, limit, offset int) (*openapi_client.UsageGetUsageSummaryLandscape200Response, error) {
	req := c.api.UsageAPI.UsageGetUsageSummaryLandscape(c.ctx, float32(teamId)).
		BeginDate(beginDate).
		EndDate(endDate)
	if limit > 0 {
		req = req.Limit(limit)
	}
	if offset > 0 {
		req = req.Offset(offset)
	}
	res, r, err := req.Execute()
	return res, errors.FormatAPIError(r, err)
}
