package dinding

// FullValidConfigForTesting is a string representation of a JSON object that contains all fields supported by the notifier Config. It can be used without secrets.
const FullValidConfigForTesting = `{
	"url": "http://localhost",
	"msgType": "actionCard",
	"title": "Alerts firing: {{ len .Alerts.Firing }}",
	"message": "{{ len .Alerts.Firing }} alerts are firing, {{ len .Alerts.Resolved }} are resolved",
	"at": {
		"atMobiles": ["1234567890", "0987654321"],
		"atUserIds": ["user1", "user2"],
		"isAtAll": true
	}
}`
