package nested

type Outer struct {
	Inner struct {
		Value int
	}
}

type A struct {
	B struct {
		C struct {
			D struct {
				E int
			}
		}
	}
}
