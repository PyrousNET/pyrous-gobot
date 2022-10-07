package wavinghands

type (
	Gesture struct {
		Name  string
		Value uint
	}

	Spell struct {
		Gestures []Gesture
	}
)
