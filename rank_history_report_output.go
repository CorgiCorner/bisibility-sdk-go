package bisibility

// Version and record-kind discriminants remain fixed; output metadata may evolve.
func rankOutputSchema(s reportSchema, field string) reportSchema {
	if field != "version" && field != "kind" {
		s.Enum = nil
		s.Const = nil
	}
	for name, child := range s.Properties {
		s.Properties[name] = rankOutputSchema(child, name)
	}
	if s.Items != nil {
		child := rankOutputSchema(*s.Items, field)
		s.Items = &child
	}
	for i, child := range s.AnyOf {
		s.AnyOf[i] = rankOutputSchema(child, field)
	}
	for i, child := range s.OneOf {
		s.OneOf[i] = rankOutputSchema(child, field)
	}
	return s
}
