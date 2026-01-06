package zenoh

import "strings"

// matchKeyExpr checks if pattern matches subject.
// Supports * (single chunk) and ** (any chunks) wildcards.
func matchKeyExpr(pattern, subject KeyExpr) bool {
	p := string(pattern)
	s := string(subject)

	// Exact match
	if p == s {
		return true
	}

	// ** matches everything
	if p == "**" {
		return true
	}

	// Simple wildcard matching
	pParts := strings.Split(p, "/")
	sParts := strings.Split(s, "/")

	return matchParts(pParts, sParts)
}

func matchParts(pattern, subject []string) bool {
	pi, si := 0, 0

	for pi < len(pattern) && si < len(subject) {
		p := pattern[pi]

		switch p {
		case "**":
			// ** at end matches everything
			if pi == len(pattern)-1 {
				return true
			}
			// Try matching rest of pattern at each position
			for i := si; i <= len(subject); i++ {
				if matchParts(pattern[pi+1:], subject[i:]) {
					return true
				}
			}
			return false
		case "*":
			// * matches single chunk
			pi++
			si++
		default:
			// Exact match required
			if p != subject[si] {
				return false
			}
			pi++
			si++
		}
	}

	// Check if both exhausted
	return pi == len(pattern) && si == len(subject)
}

