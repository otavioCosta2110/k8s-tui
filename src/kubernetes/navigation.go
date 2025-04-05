package kubernetes

type Stack struct {
	screens []ResourceInterface
}

func (s *Stack) Push(screen ResourceInterface) {
	s.screens = append(s.screens, screen)
}

func (s *Stack) Pop() ResourceInterface {
	if len(s.screens) == 0 {
		return nil
	}
	screen := s.screens[len(s.screens)-1]
	s.screens = s.screens[:len(s.screens)-1]
	return screen
}

func (s *Stack) Peek() ResourceInterface {
	if len(s.screens) == 0 {
		return nil
	}
	return s.screens[len(s.screens)-1]
}

func (s *Stack) Size() int {
	return len(s.screens)
}

var GlobalStack = Stack{}
