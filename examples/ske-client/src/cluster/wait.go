package cluster

import (
	"context"
	"net/http"

	"github.com/SchwarzIT/community-stackit-go-client/pkg/wait"
)

func (r CreateOrUpdateClusterResponse) WaitHandler(ctx context.Context, c *ClientWithResponses, projectID, clusterName string) *wait.Handler {
	return wait.New(func() (res interface{}, done bool, err error) {

		resp, err := c.GetClusterWithResponse(ctx, projectID, clusterName)
		if err != nil {
			return nil, false, err
		}
		if resp.HasError != nil {
			return nil, false, resp.HasError
		}

		status := *resp.JSON200.Status.Aggregated
		if status == STATE_HEALTHY || status == STATE_HIBERNATED {
			return resp, true, nil
		}
		return resp, false, nil
	})
}

func (r DeleteClusterResponse) WaitHandler(ctx context.Context, c *ClientWithResponses, projectID, clusterName string) *wait.Handler {
	return wait.New(func() (res interface{}, done bool, err error) {
		resp, err := c.GetClusterWithResponse(ctx, projectID, clusterName)
		if err != nil {
			return nil, false, err
		}
		if resp.HasError != nil {
			if resp.StatusCode() == http.StatusNotFound {
				return nil, true, nil
			}
			return nil, false, err
		}
		return nil, false, nil
	})
}
