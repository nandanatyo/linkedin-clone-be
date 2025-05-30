package middleware

import (
	"crypto/subtle"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/response"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SecurityConfig struct {
	RateLimitEnabled       bool
	CSRFEnabled            bool
	SecureHeadersEnabled   bool
	InputValidationEnabled bool
	SQLInjectionProtection bool
	XSSProtection          bool
}

func EnhancedSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("X-Content-Type-Options", "nosniff")

		c.Header("X-Frame-Options", "DENY")

		c.Header("X-XSS-Protection", "1; mode=block")

		c.Header("X-Powered-By", "")
		c.Header("Server", "")

		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data: https:; "+
				"font-src 'self'; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none'")

		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Header("Permissions-Policy",
			"geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()")

		c.Next()
	}
}

func SQLInjectionProtection(logger logger.Logger) gin.HandlerFunc {

	sqlPatterns := []string{
		`(?i)(union\s+select)`,
		`(?i)(select\s+.*\s+from)`,
		`(?i)(insert\s+into)`,
		`(?i)(delete\s+from)`,
		`(?i)(update\s+.*\s+set)`,
		`(?i)(drop\s+table)`,
		`(?i)(create\s+table)`,
		`(?i)(alter\s+table)`,
		`(?i)(';\s*--)`,
		`(?i)(';?\s*\/\*.*\*\/)`,
		`(?i)(0x[0-9a-f]+)`,
		`(?i)(char\s*\(\s*\d+\s*\))`,
		`(?i)(concat\s*\()`,
		`(?i)(substring\s*\()`,
		`(?i)(benchmark\s*\()`,
		`(?i)(sleep\s*\()`,
		`(?i)(waitfor\s+delay)`,
	}

	compiledPatterns := make([]*regexp.Regexp, len(sqlPatterns))
	for i, pattern := range sqlPatterns {
		compiledPatterns[i] = regexp.MustCompile(pattern)
	}

	return func(c *gin.Context) {

		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsSQLInjection(value, compiledPatterns) {
					logger.Warn("SQL injection attempt detected", map[string]interface{}{
						"ip":         c.ClientIP(),
						"user_agent": c.Request.UserAgent(),
						"path":       c.Request.URL.Path,
						"parameter":  key,
						"value":      value,
						"method":     c.Request.Method,
					})
					response.Error(c, http.StatusBadRequest, "Invalid request parameters", "Malicious input detected")
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

func XSSProtection(logger logger.Logger) gin.HandlerFunc {

	xssPatterns := []string{
		`(?i)<script[^>]*>.*?</script>`,
		`(?i)<iframe[^>]*>.*?</iframe>`,
		`(?i)<object[^>]*>.*?</object>`,
		`(?i)<embed[^>]*>`,
		`(?i)<link[^>]*>`,
		`(?i)<meta[^>]*>`,
		`(?i)javascript:`,
		`(?i)vbscript:`,
		`(?i)onload\s*=`,
		`(?i)onerror\s*=`,
		`(?i)onclick\s*=`,
		`(?i)onmouseover\s*=`,
		`(?i)onfocus\s*=`,
		`(?i)onblur\s*=`,
		`(?i)onchange\s*=`,
		`(?i)onsubmit\s*=`,
		`(?i)eval\s*\(`,
		`(?i)expression\s*\(`,
	}

	compiledPatterns := make([]*regexp.Regexp, len(xssPatterns))
	for i, pattern := range xssPatterns {
		compiledPatterns[i] = regexp.MustCompile(pattern)
	}

	return func(c *gin.Context) {

		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsXSS(value, compiledPatterns) {
					logger.Warn("XSS attempt detected", map[string]interface{}{
						"ip":         c.ClientIP(),
						"user_agent": c.Request.UserAgent(),
						"path":       c.Request.URL.Path,
						"parameter":  key,
						"value":      value,
						"method":     c.Request.Method,
					})
					response.Error(c, http.StatusBadRequest, "Invalid request parameters", "Malicious input detected")
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

func InputValidation(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		for _, param := range c.Params {
			if isInvalidInput(param.Value) {
				logger.Warn("Invalid path parameter detected", map[string]interface{}{
					"ip":         c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
					"path":       c.Request.URL.Path,
					"parameter":  param.Key,
					"value":      param.Value,
					"method":     c.Request.Method,
				})
				response.Error(c, http.StatusBadRequest, "Invalid path parameter", "")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func CSRFProtection(secret string, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			token = c.GetHeader("X-XSRF-Token")
		}

		expectedToken, err := c.Cookie("csrf-token")
		if err != nil {
			logger.Warn("CSRF token missing", map[string]interface{}{
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			})
			response.Error(c, http.StatusForbidden, "CSRF token required", "")
			c.Abort()
			return
		}

		if !isValidCSRFToken(token, expectedToken) {
			logger.Warn("CSRF token mismatch", map[string]interface{}{
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			})
			response.Error(c, http.StatusForbidden, "Invalid CSRF token", "")
			c.Abort()
			return
		}

		c.Next()
	}
}

func containsSQLInjection(input string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

func containsXSS(input string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

func isInvalidInput(input string) bool {

	if strings.Contains(input, "\x00") {
		return true
	}

	if len(input) > 1000 {
		return true
	}

	if !isValidUTF8(input) {
		return true
	}

	return false
}

func isValidUTF8(s string) bool {
	for _, r := range s {
		if r == '\uFFFD' {
			return false
		}
	}
	return true
}

func isValidCSRFToken(token, expectedToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1
}

func SecurityMonitoring(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		userAgent := c.Request.UserAgent()
		if isSuspiciousUserAgent(userAgent) {
			logger.Warn("Suspicious user agent detected", map[string]interface{}{
				"ip":         c.ClientIP(),
				"user_agent": userAgent,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			})
		}
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if status >= 400 {
			logger.Warn("Security event", map[string]interface{}{
				"ip":         c.ClientIP(),
				"user_agent": userAgent,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"status":     status,
				"latency":    latency.String(),
			})
		}
	}
}

func isSuspiciousUserAgent(userAgent string) bool {
	suspiciousPatterns := []string{
		"sqlmap", "nikto", "nmap", "masscan", "burp", "zap",
		"curl", "wget", "python-requests", "go-http-client",
	}

	ua := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(ua, pattern) {
			return true
		}
	}
	return false
}
