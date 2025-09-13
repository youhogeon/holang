package interpreter

type Environment struct {
	enclosing *Environment

	Values map[string]any
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		Values:    make(map[string]any),
		enclosing: enclosing,
	}
}

func (e *Environment) Define(name string, value any) {
	e.Values[name] = value
}

func (e *Environment) findDefinition(name string) *Environment {
	env := e

	for env != nil {
		if _, ok := env.Values[name]; ok {
			return env
		}

		env = env.enclosing
	}

	return nil
}

func (e *Environment) Assign(name string, value any) error {
	env := e.findDefinition(name)

	if env == nil {
		return NewRuntimeErrorWithLog("cannot assign to undefined variable: " + name)
	}

	env.Values[name] = value

	return nil
}

func (e *Environment) Get(name string) (any, error) {
	env := e.findDefinition(name)

	if env == nil {
		return nil, NewRuntimeErrorWithLog("undefined variable: " + name)
	}

	return env.Values[name], nil
}

func (e *Environment) GetAt(distance int, name string) (any, error) {
	env := e

	for d := 0; d < distance; d++ {
		env = env.enclosing
	}

	return env.Get(name)
}

func (e *Environment) AssignAt(distance int, name string, value any) error {
	env := e

	for d := 0; d < distance; d++ {
		env = env.enclosing
	}

	_, ok := env.Values[name]
	if !ok {
		return NewRuntimeErrorWithLog("cannot assign to undefined variable: " + name)
	}

	env.Values[name] = value

	return nil
}
