# bi-events-go
Library to instrument BI events in go services.

The following events should be logged in the given format.

The json schema of the events is as described below:

```json
{
  "$schema": "http://data4life.care/bi/events.json",
  "definitions": {
       "event": {
            "type": "object",
            "properties": {
                "event-type": {
                    "description": "The unique identifier for bi-events",
                    "type": "string"
                },
                "hostname": {
                    "description": " The content of the $HOSTNAME variable to be found in the service environment.",
                    "type": "string"
                },
                "service-name": {
                    "description": "The name of the service creating the event",
                    "type": "string"
                },
                "service-version": {
                    "description": "A unique identifier for the running service version (e.g. a GIT commit hash or tag).",
                    "type": "string"
                },
                "timestamp": {
                    "description": "An RFC3339 timestamp with at least millisecond resolution.",
                    "type": "datetime"
                },
                "user-id": {
                    "description": "The UUID (if available) of the user performing the activity",
                    "type": "string"
                },
                "activity-type": {
                    "description": "This parameter is the primary focus for analytics. The values can be login, regotsrtion, email-verification, phone-validation etc.",
                    "type": "string",
                    "enum":["login", "registration", "email-verify", "phone-verify", "lab-order-read", "lab-order-create", "account-delete", "lab-result-receive", "records-read", "records-create" ]
                },
                "data": {
                    "description": "The data format specific to activity type.",
                    "type": "object",
                    "oneOf" : [
                        {
                            "type": "object",
                            "properties": {
                                "cuc": {
                                    "description": "The Citizen Use Case of the registered user.",
                                    "type": "string"
                                },
                                "account-type": {
                                    "description": "The account type is established by emailID. Email IDs belonging to data4life.care and other internal partners are marked as internal",
                                    "type": "string",
                                    "enum": ["internal", "external"]
                                },
                                "source-url": {
                                    "description": "The feature URL that was used to onboard the user to data4life",
                                    "type": "string"
                                }
                            }
                        }
                    ]
                }
            },
            "dependencies": {
                "data": ["activity-type"]
            },
            "required": [ "event-type", "hostname", "timestamp", "activity-type", "service-name", "service-version" ]
       }
   }
}
```

Example events:

Login:

```json
{
    "event-type": "bi-event",
    "hostname": "vega123",
    "service-name": "vega",
    "service-version": "v1.2.0",
    "timestamp": "",
    "user-id": "2541c632-7cfa-4772-9b51-4c74a7618b23",
    "activity-type": "login",
    "data": {
        "source_url": "https://app.data4life.care/labres",
        "account_type": "external",
    }
}
```

Registration:

```json
{
    "event-type": "bi-event",
    "hostname": "vega123",
    "service-name": "vega",
    "service-version": "v1.2.0",
    "timestamp": "",
    "user-id": "2541c632-7cfa-4772-9b51-4c74a7618b23",
    "activity-type": "registration",
    "data": {
        "source_url": "https://app.data4life.care/labres",
        "account_type": "external",
        "cuc": "",
    }
}
```

Email Verification:

```json
{
    "event-type": "bi-event",
    "hostname": "vega123",
    "service-name": "vega",
    "service-version": "v1.2.0",
    "timestamp": "",
    "user-id": "2541c632-7cfa-4772-9b51-4c74a7618b23",
    "activity-type": "email_verification"
}
```

Phone Validation

```json
{
    "event-type": "bi-event",
    "hostname": "vega123",
    "service-name": "vega",
    "service-version": "v1.2.0",
    "timestamp": "",
    "user-id": "2541c632-7cfa-4772-9b51-4c74a7618b23",
    "activity-type": "phone_validation"
}
```
