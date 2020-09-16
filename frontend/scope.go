package frontend

type scope struct {
	dict     map[string]Type
	previous *scope
}

func newScope() *scope {
	return &scope{make(map[string]Type), nil}
}

func (s *scope) assign(name string, ty Type) {

	s.dict[name] = ty
}

func (s *scope) get(name string) Type {
	if s == nil {
		return nil
	}
	ty, ok := s.dict[name]
	if !ok {
		return s.previous.get(name)
	}
	return ty
}

func (s *scope) newScope() *scope {
	return &scope{make(map[string]Type), s}
}
