package server

import "strings"

// Label keeps one piece of information about a single server
type Label string

func (l Label) String() string {
	return string(l)
}

// GetPart returns specific part of the label, if part index is higher than last available index it returns empty string.
func (l Label) GetPart(idx int) string {
	parts := strings.Split(l.String(), ":")

	if idx < 0 {
		return ""
	}
	if len(parts) >= idx {
		return ""
	}

	return parts[idx]
}

// GetPart1 is exactly same as GetPart but it splits the label only once, this is good for IPv6 addresses
func (l Label) GetPart1(idx int) string {
	parts := strings.SplitN(l.String(), ":", 2)

	if idx < 0 {
		return ""
	}
	if len(parts) >= idx {
		return ""
	}

	return parts[idx]
}

// Labels stores multiple Label records
type Labels []Label

// StringSlice return slice of Label as strings
func (l *Labels) StringSlice() []string {
	labelsString := []string{}

	for _, label := range *l {
		labelsString = append(labelsString, label.String())
	}

	return labelsString
}
