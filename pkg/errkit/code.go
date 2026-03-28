package errkit

// Code is a scoped error identifier for classification.
type Code struct {
	scope  string
	number string
}

// Scope returns the scope of the code.
func (c Code) Scope() string { return c.scope }

// Number returns the number of the code.
func (c Code) Number() string { return c.number }

// NewCode creates a new error code with the given scope and number.
func NewCode(scope, number string) Code {
	return Code{scope: scope, number: number}
}

// String returns the code in "scope(number)" format.
func (c Code) String() string {
	if c.scope == "" && c.number == "" {
		return ""
	}
	if c.scope == "" {
		return c.number
	}
	if c.number == "" {
		return c.scope
	}

	return c.scope + "(" + c.number + ")"
}

// IsZero reports whether the code has no scope or number.
func (c Code) IsZero() bool {
	return c.scope == "" && c.number == ""
}

func (c Code) apply(e *richError) { e.code = c }
