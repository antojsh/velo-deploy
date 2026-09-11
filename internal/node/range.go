package node

import (
	"regexp"
	"strconv"
	"strings"
)

type version struct {
	major int
	minor int
	patch int
}

type bound struct {
	v    version
	open bool
	inf  bool
}

var comparatorRe = regexp.MustCompile(`(?i)(\^|~|>=|<=|>|<|=)?\s*v?(\d+)(?:\.(x|\*|\d+))?(?:\.(x|\*|\d+))?`)

func ResolveEnginesRange(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "*" {
		return DefaultNodeVersion
	}
	for _, major := range SupportedMajors {
		if rangeAllowsMajor(raw, major) {
			return strconv.Itoa(major)
		}
	}
	return DefaultNodeVersion
}

func rangeAllowsMajor(expr string, major int) bool {
	for _, orPart := range strings.Split(expr, "||") {
		if constraintsAllowMajor(strings.TrimSpace(orPart), major) {
			return true
		}
	}
	return false
}

func constraintsAllowMajor(expr string, major int) bool {
	if expr == "" {
		return true
	}
	matches := comparatorRe.FindAllStringSubmatch(expr, -1)
	if len(matches) == 0 {
		return false
	}
	min := bound{inf: true}
	max := bound{inf: true}
	for _, m := range matches {
		op := m[1]
		maj, _ := strconv.Atoi(m[2])
		minor, minorWild := parsePart(m[3])
		patch, patchWild := parsePart(m[4])
		v := version{maj, minor, patch}
		if op == "" {
			op = "="
		}
		switch op {
		case ">=":
			min = tighterMin(min, bound{v: v})
		case ">":
			min = tighterMin(min, bound{v: v, open: true})
		case "<=":
			max = tighterMax(max, bound{v: nextPatch(v)})
		case "<":
			max = tighterMax(max, bound{v: v})
		case "=":
			if minorWild || m[3] == "" {
				min = tighterMin(min, bound{v: version{maj, 0, 0}})
				max = tighterMax(max, bound{v: version{maj + 1, 0, 0}})
			} else if patchWild || m[4] == "" {
				min = tighterMin(min, bound{v: version{maj, minor, 0}})
				max = tighterMax(max, bound{v: version{maj, minor + 1, 0}})
			} else {
				min = tighterMin(min, bound{v: v})
				max = tighterMax(max, bound{v: nextPatch(v)})
			}
		case "^":
			min = tighterMin(min, bound{v: v})
			if maj > 0 {
				max = tighterMax(max, bound{v: version{maj + 1, 0, 0}})
			} else if !minorWild && minor > 0 {
				max = tighterMax(max, bound{v: version{0, minor + 1, 0}})
			} else {
				max = tighterMax(max, bound{v: version{0, 0, patch + 1}})
			}
		case "~":
			min = tighterMin(min, bound{v: v})
			if minorWild || m[3] == "" {
				max = tighterMax(max, bound{v: version{maj + 1, 0, 0}})
			} else {
				max = tighterMax(max, bound{v: version{maj, minor + 1, 0}})
			}
		}
	}
	low := version{major, 0, 0}
	high := version{major + 1, 0, 0}
	return intervalsOverlap(low, high, min, max)
}

func parsePart(s string) (int, bool) {
	if s == "" || strings.EqualFold(s, "x") || s == "*" {
		return 0, true
	}
	n, _ := strconv.Atoi(s)
	return n, false
}

func nextPatch(v version) version {
	return version{v.major, v.minor, v.patch + 1}
}

func cmpVer(a, b version) int {
	if a.major != b.major {
		return a.major - b.major
	}
	if a.minor != b.minor {
		return a.minor - b.minor
	}
	return a.patch - b.patch
}

func tighterMin(current, incoming bound) bound {
	if current.inf {
		return incoming
	}
	c := cmpVer(incoming.v, current.v)
	if c > 0 || (c == 0 && incoming.open && !current.open) {
		return incoming
	}
	return current
}

func tighterMax(current, incoming bound) bound {
	if current.inf {
		return incoming
	}
	c := cmpVer(incoming.v, current.v)
	if c < 0 || (c == 0 && incoming.open && !current.open) {
		return incoming
	}
	return current
}

func intervalsOverlap(low, high version, minB, maxB bound) bool {
	min := low
	if !minB.inf {
		min = minB.v
		if minB.open {
			min = nextPatch(minB.v)
		}
	}
	max := version{9999, 0, 0}
	if !maxB.inf {
		max = maxB.v
	}
	return cmpVer(low, max) < 0 && cmpVer(min, high) < 0
}
