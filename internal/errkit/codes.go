package errkit

// Scope builds error codes within a namespace.
type Scope struct {
	path string
}

// Code creates a new code in this scope with the given number.
func (s Scope) Code(number string) Code {
	return NewCode(s.path, number)
}

// Scope creates a child scope with the given name.
func (s Scope) Scope(name string) Scope {
	if s.path == "" {
		return Scope{path: name}
	}

	return Scope{path: s.path + "." + name}
}

// Predefined error scopes for common error categories.
var (
	User = UserScope{
		Scope:       Scope{"user"},
		Request:     Scope{"user.request"},
		Auth:        Scope{"user.auth"},
		Forbidden:   Scope{"user.forbidden"},
		Permissions: Scope{"user.permissions"},
		RateLimit:   Scope{"user.rate_limit"},
	}

	App = AppScope{
		Scope:       Scope{"app"},
		Internal:    Scope{"app.internal"},
		Validation:  Scope{"app.validation"},
		DB:          Scope{"app.db"},
		Resource:    Scope{"app.resource"},
		Dependency:  Scope{"app.dependency"},
		Unavailable: Scope{"app.unavailable"},
	}

	System = SystemScope{
		Scope:       Scope{"system"},
		Timeout:     Scope{"system.timeout"},
		Unavailable: Scope{"system.unavailable"},
	}
)

// UserScope contains user-related error scopes.
type UserScope struct {
	Scope
	Request     Scope
	Auth        Scope
	Forbidden   Scope
	Permissions Scope
	RateLimit   Scope
}

// AppScope contains application-related error scopes.
type AppScope struct {
	Scope
	Internal    Scope
	Validation  Scope
	DB          Scope
	Resource    Scope
	Dependency  Scope
	Unavailable Scope
}

// SystemScope contains system-level error scopes.
type SystemScope struct {
	Scope
	Timeout     Scope
	Unavailable Scope
}
