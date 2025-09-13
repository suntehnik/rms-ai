package docs

// Authorization Patterns and Error Response Documentation
// This file contains comprehensive documentation for authorization patterns,
// role-based access control, and security error responses used throughout the API.

// RoleHierarchy represents the role hierarchy in the system
// @Description Role hierarchy showing permission levels from highest to lowest
type RoleHierarchy struct {
	Administrator struct {
		Level        int      `json:"level" example:"3"`
		Description  string   `json:"description" example:"Highest privilege level with full system access"`
		Capabilities []string `json:"capabilities" example:"[\"user_management\",\"system_configuration\",\"all_entity_operations\",\"administrative_functions\"]"`
	} `json:"administrator"`
	User struct {
		Level        int      `json:"level" example:"2"`
		Description  string   `json:"description" example:"Standard user with entity management capabilities"`
		Capabilities []string `json:"capabilities" example:"[\"create_entities\",\"edit_entities\",\"delete_entities\",\"view_entities\",\"comment_system\"]"`
	} `json:"user"`
	Commenter struct {
		Level        int      `json:"level" example:"1"`
		Description  string   `json:"description" example:"Limited user with commenting and viewing capabilities"`
		Capabilities []string `json:"capabilities" example:"[\"view_entities\",\"create_comments\",\"edit_own_comments\",\"resolve_comments\"]"`
	} `json:"commenter"`
} // @name RoleHierarchy

// AuthorizationMatrix represents the complete authorization matrix for all operations
// @Description Complete authorization matrix showing which roles can perform which operations
type AuthorizationMatrix struct {
	EntityOperations struct {
		CreateEpic struct {
			RequiredRole  string `json:"required_role" example:"User"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"true"`
			Commenter     bool   `json:"commenter" example:"false"`
			Description   string `json:"description" example:"Create new epics in the system"`
		} `json:"create_epic"`
		EditEpic struct {
			RequiredRole  string `json:"required_role" example:"User"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"true"`
			Commenter     bool   `json:"commenter" example:"false"`
			Description   string `json:"description" example:"Modify existing epic properties"`
		} `json:"edit_epic"`
		DeleteEpic struct {
			RequiredRole  string `json:"required_role" example:"User"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"true"`
			Commenter     bool   `json:"commenter" example:"false"`
			Description   string `json:"description" example:"Delete epics and their associated data"`
		} `json:"delete_epic"`
		ViewEpic struct {
			RequiredRole  string `json:"required_role" example:"Commenter"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"true"`
			Commenter     bool   `json:"commenter" example:"true"`
			Description   string `json:"description" example:"View epic details and associated entities"`
		} `json:"view_epic"`
	} `json:"entity_operations"`
	UserManagement struct {
		CreateUser struct {
			RequiredRole  string `json:"required_role" example:"Administrator"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"false"`
			Commenter     bool   `json:"commenter" example:"false"`
			Description   string `json:"description" example:"Create new user accounts"`
		} `json:"create_user"`
		EditUser struct {
			RequiredRole  string `json:"required_role" example:"Administrator"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"false"`
			Commenter     bool   `json:"commenter" example:"false"`
			Description   string `json:"description" example:"Modify user account details and roles"`
		} `json:"edit_user"`
		DeleteUser struct {
			RequiredRole  string `json:"required_role" example:"Administrator"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"false"`
			Commenter     bool   `json:"commenter" example:"false"`
			Description   string `json:"description" example:"Delete user accounts"`
		} `json:"delete_user"`
		ViewUsers struct {
			RequiredRole  string `json:"required_role" example:"Administrator"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"false"`
			Commenter     bool   `json:"commenter" example:"false"`
			Description   string `json:"description" example:"View all user accounts and details"`
		} `json:"view_users"`
	} `json:"user_management"`
	SystemConfiguration struct {
		ManageRequirementTypes struct {
			RequiredRole  string `json:"required_role" example:"Administrator"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"false"`
			Commenter     bool   `json:"commenter" example:"false"`
			Description   string `json:"description" example:"Create, edit, and delete requirement types"`
		} `json:"manage_requirement_types"`
		ManageRelationshipTypes struct {
			RequiredRole  string `json:"required_role" example:"Administrator"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"false"`
			Commenter     bool   `json:"commenter" example:"false"`
			Description   string `json:"description" example:"Create, edit, and delete relationship types"`
		} `json:"manage_relationship_types"`
		ManageStatusModels struct {
			RequiredRole  string `json:"required_role" example:"Administrator"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"false"`
			Commenter     bool   `json:"commenter" example:"false"`
			Description   string `json:"description" example:"Configure status models and transitions"`
		} `json:"manage_status_models"`
	} `json:"system_configuration"`
	CommentSystem struct {
		CreateComment struct {
			RequiredRole  string `json:"required_role" example:"Commenter"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"true"`
			Commenter     bool   `json:"commenter" example:"true"`
			Description   string `json:"description" example:"Create comments on entities"`
		} `json:"create_comment"`
		EditOwnComment struct {
			RequiredRole  string `json:"required_role" example:"Commenter"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"true"`
			Commenter     bool   `json:"commenter" example:"true"`
			Description   string `json:"description" example:"Edit comments authored by the user"`
		} `json:"edit_own_comment"`
		EditAnyComment struct {
			RequiredRole  string `json:"required_role" example:"User"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"true"`
			Commenter     bool   `json:"commenter" example:"false"`
			Description   string `json:"description" example:"Edit any comment in the system"`
		} `json:"edit_any_comment"`
		ResolveComment struct {
			RequiredRole  string `json:"required_role" example:"Commenter"`
			Administrator bool   `json:"administrator" example:"true"`
			User          bool   `json:"user" example:"true"`
			Commenter     bool   `json:"commenter" example:"true"`
			Description   string `json:"description" example:"Mark comments as resolved or unresolved"`
		} `json:"resolve_comment"`
	} `json:"comment_system"`
} // @name AuthorizationMatrix

// AuthorizationErrorTypes represents different types of authorization errors
// @Description Comprehensive list of authorization error types and their meanings
type AuthorizationErrorTypes struct {
	AuthenticationRequired struct {
		HTTPStatus  int    `json:"http_status" example:"401"`
		ErrorCode   string `json:"error_code" example:"AUTHENTICATION_REQUIRED"`
		Message     string `json:"message" example:"Authentication required"`
		Description string `json:"description" example:"Request requires valid JWT token in Authorization header"`
		Example     string `json:"example" example:"Missing or empty Authorization header"`
	} `json:"authentication_required"`
	InvalidToken struct {
		HTTPStatus  int    `json:"http_status" example:"401"`
		ErrorCode   string `json:"error_code" example:"INVALID_TOKEN"`
		Message     string `json:"message" example:"Invalid token"`
		Description string `json:"description" example:"JWT token is malformed, corrupted, or has invalid signature"`
		Example     string `json:"example" example:"Malformed JWT token or invalid signature"`
	} `json:"invalid_token"`
	TokenExpired struct {
		HTTPStatus  int    `json:"http_status" example:"401"`
		ErrorCode   string `json:"error_code" example:"TOKEN_EXPIRED"`
		Message     string `json:"message" example:"Token expired"`
		Description string `json:"description" example:"JWT token has passed its expiration time"`
		Example     string `json:"example" example:"Token expired at 2023-01-01T12:00:00Z"`
	} `json:"token_expired"`
	InsufficientPermissions struct {
		HTTPStatus  int    `json:"http_status" example:"403"`
		ErrorCode   string `json:"error_code" example:"INSUFFICIENT_PERMISSIONS"`
		Message     string `json:"message" example:"Insufficient permissions"`
		Description string `json:"description" example:"User role does not have required permissions for this operation"`
		Example     string `json:"example" example:"Commenter role cannot create epics (User role required)"`
	} `json:"insufficient_permissions"`
	RoleRequired struct {
		HTTPStatus  int    `json:"http_status" example:"403"`
		ErrorCode   string `json:"error_code" example:"ROLE_REQUIRED"`
		Message     string `json:"message" example:"Administrator role required"`
		Description string `json:"description" example:"Operation requires specific role level"`
		Example     string `json:"example" example:"User management operations require Administrator role"`
	} `json:"role_required"`
} // @name AuthorizationErrorTypes

// SecurityMiddlewareFlow represents the security middleware processing flow
// @Description Step-by-step flow of security middleware processing for authenticated requests
type SecurityMiddlewareFlow struct {
	Step1 struct {
		Name        string `json:"name" example:"Header Extraction"`
		Description string `json:"description" example:"Extract Authorization header from HTTP request"`
		Success     string `json:"success" example:"Authorization header found with Bearer prefix"`
		Failure     string `json:"failure" example:"Return 401 - Authorization header required"`
	} `json:"step1"`
	Step2 struct {
		Name        string `json:"name" example:"Token Validation"`
		Description string `json:"description" example:"Validate JWT token signature and structure"`
		Success     string `json:"success" example:"Token is valid and properly signed"`
		Failure     string `json:"failure" example:"Return 401 - Invalid token"`
	} `json:"step2"`
	Step3 struct {
		Name        string `json:"name" example:"Expiration Check"`
		Description string `json:"description" example:"Verify token has not expired"`
		Success     string `json:"success" example:"Token is within valid time range"`
		Failure     string `json:"failure" example:"Return 401 - Token expired"`
	} `json:"step3"`
	Step4 struct {
		Name        string `json:"name" example:"Claims Extraction"`
		Description string `json:"description" example:"Extract user claims from validated token"`
		Success     string `json:"success" example:"User ID, username, and role extracted successfully"`
		Failure     string `json:"failure" example:"Return 500 - Invalid claims structure"`
	} `json:"step4"`
	Step5 struct {
		Name        string `json:"name" example:"Role Authorization"`
		Description string `json:"description" example:"Check if user role has required permissions"`
		Success     string `json:"success" example:"User role meets minimum requirements"`
		Failure     string `json:"failure" example:"Return 403 - Insufficient permissions"`
	} `json:"step5"`
	Step6 struct {
		Name        string `json:"name" example:"Context Storage"`
		Description string `json:"description" example:"Store user claims in request context"`
		Success     string `json:"success" example:"Claims available to handler functions"`
		Failure     string `json:"failure" example:"N/A - Process continues to handler"`
	} `json:"step6"`
} // @name SecurityMiddlewareFlow

// TokenManagementBestPractices represents best practices for JWT token management
// @Description Comprehensive best practices for secure JWT token management
type TokenManagementBestPractices struct {
	ClientSide struct {
		Storage struct {
			Recommended string   `json:"recommended" example:"Secure httpOnly cookies or encrypted localStorage"`
			Avoid       []string `json:"avoid" example:"[\"Plain localStorage\",\"sessionStorage\",\"URL parameters\",\"Local variables\"]"`
			Reasoning   string   `json:"reasoning" example:"Prevent XSS attacks and token theft"`
		} `json:"storage"`
		Transmission struct {
			Protocol   string `json:"protocol" example:"HTTPS only"`
			HeaderName string `json:"header_name" example:"Authorization"`
			Format     string `json:"format" example:"Bearer <token>"`
			Validation string `json:"validation" example:"Always validate token format before sending"`
		} `json:"transmission"`
		ErrorHandling struct {
			ExpiredToken string `json:"expired_token" example:"Automatically attempt token refresh or redirect to login"`
			InvalidToken string `json:"invalid_token" example:"Clear stored token and redirect to login"`
			NetworkError string `json:"network_error" example:"Retry with exponential backoff"`
			ServerError  string `json:"server_error" example:"Show user-friendly error message"`
		} `json:"error_handling"`
	} `json:"client_side"`
	ServerSide struct {
		TokenGeneration struct {
			Algorithm  string `json:"algorithm" example:"HS256 (HMAC with SHA-256)"`
			SecretKey  string `json:"secret_key" example:"Use strong, randomly generated secret key"`
			Expiration string `json:"expiration" example:"Short-lived tokens (1 hour recommended)"`
			Claims     string `json:"claims" example:"Include minimal necessary user information"`
		} `json:"token_generation"`
		TokenValidation struct {
			SignatureCheck   string `json:"signature_check" example:"Always verify token signature"`
			ExpirationCheck  string `json:"expiration_check" example:"Reject expired tokens immediately"`
			ClaimsValidation string `json:"claims_validation" example:"Validate all required claims are present"`
			Blacklisting     string `json:"blacklisting" example:"Optional: Maintain blacklist for revoked tokens"`
		} `json:"token_validation"`
		SecurityHeaders struct {
			CORS            string `json:"cors" example:"Configure appropriate CORS policies"`
			ContentSecurity string `json:"content_security" example:"Set CSP headers to prevent XSS"`
			HSTS            string `json:"hsts" example:"Use HTTP Strict Transport Security"`
		} `json:"security_headers"`
	} `json:"server_side"`
} // @name TokenManagementBestPractices

// AuthorizationWorkflowExamples represents common authorization workflow examples
// @Description Real-world examples of authorization workflows for different scenarios
type AuthorizationWorkflowExamples struct {
	UserLogin struct {
		Step1 string `json:"step1" example:"POST /auth/login with username/password"`
		Step2 string `json:"step2" example:"Server validates credentials against database"`
		Step3 string `json:"step3" example:"Server generates JWT token with user claims"`
		Step4 string `json:"step4" example:"Server returns token and user information"`
		Step5 string `json:"step5" example:"Client stores token securely"`
		Step6 string `json:"step6" example:"Client includes token in subsequent requests"`
	} `json:"user_login"`
	CreateEpic struct {
		Step1 string `json:"step1" example:"Client sends POST /api/v1/epics with Authorization header"`
		Step2 string `json:"step2" example:"Middleware extracts and validates JWT token"`
		Step3 string `json:"step3" example:"Middleware checks user role (User or Administrator required)"`
		Step4 string `json:"step4" example:"Handler processes epic creation request"`
		Step5 string `json:"step5" example:"Server returns created epic with 201 status"`
	} `json:"create_epic"`
	AdminOperation struct {
		Step1 string `json:"step1" example:"Client sends POST /auth/users with Authorization header"`
		Step2 string `json:"step2" example:"Middleware validates JWT token"`
		Step3 string `json:"step3" example:"Middleware checks for Administrator role"`
		Step4 string `json:"step4" example:"If not Administrator, return 403 Forbidden"`
		Step5 string `json:"step5" example:"If Administrator, proceed to user creation"`
	} `json:"admin_operation"`
	TokenExpiration struct {
		Step1 string `json:"step1" example:"Client sends request with expired token"`
		Step2 string `json:"step2" example:"Middleware detects token expiration"`
		Step3 string `json:"step3" example:"Server returns 401 with TOKEN_EXPIRED error"`
		Step4 string `json:"step4" example:"Client attempts token refresh or redirects to login"`
		Step5 string `json:"step5" example:"User re-authenticates to obtain new token"`
	} `json:"token_expiration"`
} // @name AuthorizationWorkflowExamples

// SecurityAuditingGuidelines represents security auditing and monitoring guidelines
// @Description Guidelines for security auditing and monitoring of authentication/authorization
type SecurityAuditingGuidelines struct {
	LoggingRequirements struct {
		AuthenticationAttempts string `json:"authentication_attempts" example:"Log all login attempts with timestamp, IP, and result"`
		AuthorizationFailures  string `json:"authorization_failures" example:"Log all 403 Forbidden responses with user and attempted operation"`
		TokenOperations        string `json:"token_operations" example:"Log token generation, validation failures, and expiration events"`
		SuspiciousActivity     string `json:"suspicious_activity" example:"Log multiple failed attempts, unusual access patterns"`
	} `json:"logging_requirements"`
	MonitoringMetrics struct {
		AuthenticationRate string `json:"authentication_rate" example:"Track successful vs failed authentication rates"`
		TokenUsage         string `json:"token_usage" example:"Monitor token usage patterns and expiration rates"`
		RoleDistribution   string `json:"role_distribution" example:"Track distribution of user roles and permission usage"`
		ErrorRates         string `json:"error_rates" example:"Monitor 401/403 error rates and patterns"`
	} `json:"monitoring_metrics"`
	SecurityAlerts struct {
		BruteForceAttempts  string `json:"brute_force_attempts" example:"Alert on multiple failed login attempts from same IP"`
		PrivilegeEscalation string `json:"privilege_escalation" example:"Alert on attempts to access admin functions without proper role"`
		TokenAnomalies      string `json:"token_anomalies" example:"Alert on unusual token usage patterns or validation failures"`
		SystemAccess        string `json:"system_access" example:"Alert on access to sensitive configuration endpoints"`
	} `json:"security_alerts"`
} // @name SecurityAuditingGuidelines
