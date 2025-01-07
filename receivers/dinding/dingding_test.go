package dinding

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"

	"github.com/grafana/alerting/logging"
	"github.com/grafana/alerting/receivers"
	"github.com/grafana/alerting/templates"
)

func TestNotify(t *testing.T) {
	tmpl := templates.ForTests(t)

	externalURL, err := url.Parse("http://localhost")
	require.NoError(t, err)
	tmpl.ExternalURL = externalURL

	cases := []struct {
		name        string
		settings    Config
		alerts      []*types.Alert
		expMsg      map[string]interface{}
		expMsgError error
	}{
		{
			name: "Default config with one alert and at field",
			settings: Config{
				URL:         "http://localhost",
				MessageType: defaultDingdingMsgType,
				Title:       templates.DefaultMessageTitleEmbed,
				Message:     templates.DefaultMessageEmbed,
				At: At{
					AtMobiles: []string{"1234567890"},
					AtUserIds: []string{"user1", "user2"},
					IsAtAll:   true,
				},
			},
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1", "__dashboardUid__": "abcd", "__panelId__": "efgh", "__values__": "{\"A\": 1234}", "__value_string__": "1234"},
					},
				},
			},
			expMsg: map[string]interface{}{
				"msgtype": "link",
				"link": map[string]interface{}{
					"messageUrl": "dingtalk://dingtalkclient/page/link?pc_slide=false&url=http%3A%2F%2Flocalhost%2Falerting%2Flist",
					"text":       "**Firing**\n\nValue: A=1234\nLabels:\n - alertname = alert1\n - lbl1 = val1\nAnnotations:\n - ann1 = annv1\nSilence: http://localhost/alerting/silence/new?alertmanager=grafana&matcher=alertname%3Dalert1&matcher=lbl1%3Dval1\nDashboard: http://localhost/d/abcd\nPanel: http://localhost/d/abcd?viewPanel=efgh\n",
					"title":      "[FIRING:1]  (val1)",
				},
				"at": map[string]interface{}{
					"atMobiles": []string{"1234567890"},
					"atUserIds": []string{"user1", "user2"},
					"isAtAll":   true,
				},
			},
			expMsgError: nil,
		},
		{
			name: "Custom config with multiple alerts and at field",
			settings: Config{
				URL:         "http://localhost",
				MessageType: "actionCard",
				Title:       templates.DefaultMessageTitleEmbed,
				Message:     "{{ len .Alerts.Firing }} alerts are firing, {{ len .Alerts.Resolved }} are resolved",
				At: At{
					AtMobiles: []string{"1234567890", "0987654321"},
					AtUserIds: []string{"user1", "user2"},
					IsAtAll:   true,
				},
			},
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1"},
					},
				}, {
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val2"},
						Annotations: model.LabelSet{"ann1": "annv2"},
					},
				},
			},
			expMsg: map[string]interface{}{
				"actionCard": map[string]interface{}{
					"singleTitle": "More",
					"singleURL":   "dingtalk://dingtalkclient/page/link?pc_slide=false&url=http%3A%2F%2Flocalhost%2Falerting%2Flist",
					"text":        "2 alerts are firing, 0 are resolved",
					"title":       "[FIRING:2]  ",
				},
				"msgtype": "actionCard",
				"at": map[string]interface{}{
					"atMobiles": []string{"1234567890", "0987654321"},
					"atUserIds": []string{"user1", "user2"},
					"isAtAll":   true,
				},
			},
			expMsgError: nil,
		},
		{
			name: "Default config with one alert and custom title and description with at field",
			settings: Config{
				URL:         "http://localhost",
				MessageType: defaultDingdingMsgType,
				Title:       "Alerts firing: {{ len .Alerts.Firing }}",
				Message:     "customMessage",
				At: At{
					AtMobiles: []string{"1234567890"},
					AtUserIds: []string{"user1", "user2"},
					IsAtAll:   true,
				},
			},
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1", "__dashboardUid__": "abcd", "__panelId__": "efgh", "__values__": "{\"A\": 1234}", "__value_string__": "1234"},
					},
				},
			},
			expMsg: map[string]interface{}{
				"msgtype": "link",
				"link": map[string]interface{}{
					"messageUrl": "dingtalk://dingtalkclient/page/link?pc_slide=false&url=http%3A%2F%2Flocalhost%2Falerting%2Flist",
					"text":       "customMessage",
					"title":      "Alerts firing: 1",
				},
				"at": map[string]interface{}{
					"atMobiles": []string{"1234567890"},
					"atUserIds": []string{"user1", "user2"},
					"isAtAll":   true,
				},
			},
			expMsgError: nil,
		},
		{
			name: "Missing field in template with at field",
			settings: Config{
				URL:         "http://localhost",
				MessageType: "actionCard",
				Title:       templates.DefaultMessageTitleEmbed,
				Message:     "I'm a custom template {{ .NotAField }} bad template",
				At: At{
					AtMobiles: []string{"1234567890"},
					AtUserIds: []string{"user1", "user2"},
					IsAtAll:   true,
				},
			},
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1"},
					},
				}, {
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val2"},
						Annotations: model.LabelSet{"ann1": "annv2"},
					},
				},
			},
			expMsg: map[string]interface{}{
				"link": map[string]interface{}{
					"messageUrl": "dingtalk://dingtalkclient/page/link?pc_slide=false&url=http%3A%2F%2Flocalhost%2Falerting%2Flist",
					"text":       "I'm a custom template ",
					"title":      "",
				},
				"msgtype": "link",
				"at": map[string]interface{}{
					"atMobiles": []string{"1234567890"},
					"atUserIds": []string{"user1", "user2"},
					"isAtAll":   true,
				},
			},
			expMsgError: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			webhookSender := receivers.MockNotificationService()
			pn := &Notifier{
				Base: &receivers.Base{
					Name:                  "",
					Type:                  "",
					UID:                   "",
					DisableResolveMessage: false,
				},
				log:      &logging.FakeLogger{},
				ns:       webhookSender,
				tmpl:     tmpl,
				settings: c.settings,
			}

			ctx := notify.WithGroupKey(context.Background(), "alertname")
			ctx = notify.WithGroupLabels(ctx, model.LabelSet{"alertname": ""})

			ok, err := pn.Notify(ctx, c.alerts...)

			if c.expMsgError != nil {
				require.False(t, ok)
				require.Error(t, err)
				require.Equal(t, c.expMsgError.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			require.True(t, ok)

			require.NotEmpty(t, webhookSender.Webhook.URL)

			expBody, err := json.Marshal(c.expMsg)
			require.NoError(t, err)

			require.JSONEq(t, string(expBody), webhookSender.Webhook.Body)
		})
	}
}
