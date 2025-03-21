package typesystem

type IgnoreFunc func(id GIRIdentifier) bool

func ignoreOr(sf ...IgnoreFunc) IgnoreFunc {
	if len(sf) == 0 {
		return func(id GIRIdentifier) bool {
			return false
		}
	}
	if len(sf) == 1 {
		return sf[0]
	}
	return func(id GIRIdentifier) bool {
		for _, f := range sf {
			if f(id) {
				return true
			}
		}

		return false
	}
}

func IgnoreMatching(pattern GIRIdentifier) IgnoreFunc {
	return func(id GIRIdentifier) bool {
		return id.Matches(pattern)
	}
}
