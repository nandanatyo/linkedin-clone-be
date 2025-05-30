package logger

type Fields map[string]interface{}

const (
	FieldUserID     = "user_id"
	FieldRequestID  = "request_id"
	FieldIP         = "ip"
	FieldUserAgent  = "user_agent"
	FieldMethod     = "method"
	FieldPath       = "path"
	FieldStatusCode = "status_code"
	FieldLatency    = "latency"
	FieldError      = "error"
	FieldAction     = "action"
	FieldResource   = "resource"
	FieldComponent  = "component"
	FieldFunction   = "function"
)

func HTTPFields(method, path, userAgent, ip string) Fields {
	return Fields{
		FieldMethod:    method,
		FieldPath:      path,
		FieldUserAgent: userAgent,
		FieldIP:        ip,
	}
}

func UserFields(userID uint, email, username string) Fields {
	return Fields{
		FieldUserID: userID,
		"email":     email,
		"username":  username,
	}
}

func ErrorFields(err error, component, function string) Fields {
	fields := Fields{
		FieldComponent: component,
		FieldFunction:  function,
	}

	if err != nil {
		fields[FieldError] = err.Error()
	}

	return fields
}
