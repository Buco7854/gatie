package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valkey-io/valkey-go"
)

type HealthBody struct {
	Status   string            `json:"status" enum:"ok,unhealthy" doc:"Overall health status"`
	Services map[string]string `json:"services" doc:"Individual service statuses"`
}

type HealthOutput struct {
	Body HealthBody
}

func RegisterHealth(api huma.API, db *pgxpool.Pool, vk valkey.Client) {
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/api/health",
		Summary:     "Health check",
		Tags:        []string{"System"},
	}, func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
		services := make(map[string]string)
		healthy := true

		if err := db.Ping(ctx); err != nil {
			services["postgres"] = "unreachable"
			healthy = false
		} else {
			services["postgres"] = "ok"
		}

		cmd := vk.Do(ctx, vk.B().Ping().Build())
		if cmd.Error() != nil {
			services["valkey"] = "unreachable"
			healthy = false
		} else {
			services["valkey"] = "ok"
		}

		if !healthy {
			var details []error
			for svc, status := range services {
				if status != "ok" {
					details = append(details, &huma.ErrorDetail{
						Location: svc,
						Message:  status,
						Value:    status,
					})
				}
			}
			return nil, huma.NewError(http.StatusServiceUnavailable, "service unhealthy", details...)
		}

		return &HealthOutput{
			Body: HealthBody{
				Status:   "ok",
				Services: services,
			},
		}, nil
	})
}
