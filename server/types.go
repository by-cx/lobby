package server

// Label keeps one piece of information about a single server
type Label string

func (l Label) String() string {
	return string(l)
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
