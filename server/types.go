package server

// Label keeps one piece of information about a single server
type Label string

func (l Label) String() string {
	return string(l)
}

// Labels stores multiple Label records
type Labels []Label
