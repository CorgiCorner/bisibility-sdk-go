package bisibility

// Version and record-kind discriminants remain fixed; output metadata may evolve.
func rankOutputSchema(s reportSchema, field string) reportSchema {
	// Evidence was added after version 2 shipped. Accept older records without
	// the projection while still validating it, including nested required fields,
	// whenever a newer server supplies it.
	if _, ok := s.Properties["evidence_projection"]; ok {
		required := make([]string, 0, len(s.Required))
		for _, name := range s.Required {
			if name != "evidence_projection" {
				required = append(required, name)
			}
		}
		s.Required = required
	}
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
