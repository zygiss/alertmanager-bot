package alertmanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/hako/durafmt"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
)

type alertResponse struct {
	Status string        `json:"status"`
	Alerts []types.Alert `json:"data,omitempty"`
}

// ListAlerts returns a slice of Alert and an error.
func ListAlerts(logger log.Logger, alertmanagerURL string) ([]types.Alert, error) {
	resp, err := httpRetry(logger, http.MethodGet, alertmanagerURL+"/api/v1/alerts")
	if err != nil {
		return nil, err
	}

	var alertResponse alertResponse
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	if err := dec.Decode(&alertResponse); err != nil {
		return nil, err
	}

	return alertResponse.Alerts, err
}

// AlertMessage converts an alert to a message string
func AlertMessage(a types.Alert) string {
	var status, duration string
	var formattedAnnotations, formattedLabels string
	switch a.Status() {
	case model.AlertFiring:
		status = fmt.Sprintf("🔥 *%s* 🔥", strings.ToUpper(string(a.Status())))
		duration = fmt.Sprintf("*Started*: %s ago", durafmt.Parse(time.Since(a.StartsAt)))
	case model.AlertResolved:
		status = fmt.Sprintf("*%s*", strings.ToUpper(string(a.Status())))
		duration = fmt.Sprintf(
			"*Ended*: %s ago\n*Duration*: %s",
			durafmt.Parse(time.Since(a.EndsAt)),
			durafmt.Parse(a.EndsAt.Sub(a.StartsAt)),
		)
	}

	summary := a.Annotations["summary"]
	delete(a.Annotations, "summary")
	for k, v := range a.Annotations {
		formattedAnnotations = formattedAnnotations + fmt.Sprintf(
			"%s: %s\n", k, escape(v),
		)
	}

	alertName := a.Labels["alertname"]
	delete(a.Labels, "alertname")
	for k, v := range a.Labels {
		formattedLabels = formattedLabels + fmt.Sprintf(
			"%s: %s\n", k, escape(v),
		)
	}

	return fmt.Sprintf(
		"%s *%s*\n%s\n\n%s%s\n%s\n",
		status,
		alertName,
		summary,
		formattedLabels,
		formattedAnnotations,
		duration,
	)
}

func escape(s model.LabelValue) string {
	return strings.Replace(string(s), "_", "\\_", -1)
}
