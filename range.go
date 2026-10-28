package main
import (
	"strings"
	"strconv"
)

func parseParts(s string) (int, int, int, bool) {
	parts := strings.Split(s, ".")
	if (len(parts) > 3) {
		return 0, 0, 0, false
	}

	var nums [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return 0, 0, 0, false
		}
		nums[i] = n
	}
	return nums[0], nums[1], nums[2], true
}

func supportedRange(rng string) bool {
	if rng == "*" {
		return true
	}
	if strings.HasPrefix(rng, "^") {
		_, _, _, ok := parseParts(strings.TrimPrefix(rng, "^"))
		return ok
	}
	if strings.HasPrefix(rng, "~") {
		_, _, _, ok := parseParts(strings.TrimPrefix(rng, "~"))
		return ok
	}
	parts := strings.Split(rng, ".")
	_, _, _, ok := parseParts(rng)
	return len(parts) == 3 && ok
}


func rangeStatus(rng, latest string) string {
	rng, _, _ = strings.Cut(rng, "+")
	latest, _, _ = strings.Cut(latest, "+")

	if rng == "*" {
		return "inside"
	}

	if (strings.Contains(rng, "-") || strings.Contains(latest, "-")) {
		return "unparsed"
	}

	if (strings.HasPrefix(rng, "^")) {
		major, minor, patch, ok := parseParts(strings.TrimPrefix(rng, "^"))
		if !ok {
			return "unparsed"
		}
		latestMajor, latestMinor, latestPatch, ok := parseParts(latest)
		if !ok || len(strings.Split(latest, ".")) != 3 {
			return "unparsed"
		}

		below := latestMajor < major || (latestMajor == major && latestMinor < minor) || (latestMajor == major && latestMinor == minor && latestPatch < patch)
		equal := latestMajor == major && latestMinor == minor && latestPatch == patch
		above := latestMajor > major || (latestMajor == major && latestMinor > minor) || (latestMajor == major && latestMinor == minor && latestPatch > patch)
	
		if below {
			return "outside"
		} else if above {
			if (major > 0) {
				if latestMajor > major {
					return "outside"
				} else {
					return "inside"
				}
			} else {
				caratParts := len(strings.Split(strings.TrimPrefix(rng, "^"), "."))
				if caratParts == 1 {
					if latestMajor > major {
						return "outside"
					} else {
						return "inside"
					}
				} else if caratParts == 2 || (minor > 0 && caratParts == 3) {
					if latestMajor > major {
						return "outside"
					}
					if latestMinor > minor {
						return "outside"
					} else {
						return "inside"
					}
				} else if caratParts == 3 {
					return "outside"
				}
				return "unparsed"
			}
		} else if equal {
			return "inside"
		}
	}

	if (strings.HasPrefix(rng, "~")) {
		major, minor, patch, ok := parseParts(strings.TrimPrefix(rng, "~"))
		if !ok {
			return "unparsed"
		}
		latestMajor, latestMinor, latestPatch, ok := parseParts(latest)
		if !ok || len(strings.Split(latest, ".")) != 3 {
			return "unparsed"
		}
		below := latestMajor < major || (latestMajor == major && latestMinor < minor) || (latestMajor == major && latestMinor == minor && latestPatch < patch)
		equal := latestMajor == major && latestMinor == minor && latestPatch == patch
		above := latestMajor > major || (latestMajor == major && latestMinor > minor) || (latestMajor == major && latestMinor == minor && latestPatch > patch)

		if below {
			return "outside"
		} else if above {
			tildeParts := len(strings.Split(strings.TrimPrefix(rng, "~"), "."))
			if tildeParts == 1 {
				if latestMajor > major {
					return "outside"
				} else {
					return "inside"
				}
			} else if tildeParts == 2 || tildeParts == 3 {
				if latestMajor > major {
					return "outside"
				}
				if latestMinor > minor {
					return "outside"
				} else {
					return "inside"
				}
			} else {
				return "unparsed"
			}
		} else if equal {
			return "inside"
		}
	}

	rngParts := strings.Split(rng, ".")
	latestParts := strings.Split(latest, ".")

	if len(rngParts) != 3 || len(latestParts) != 3 {
		return "unparsed"
	}

	for i := range rngParts {
		rngInt, err := strconv.Atoi(rngParts[i])
		if err != nil {
			return "unparsed"
		}
		latestInt, err := strconv.Atoi(latestParts[i])
		if err != nil {
			return "unparsed"
		}
		if latestInt != rngInt {
			return "outside"
		}
	}

	return "inside"
}