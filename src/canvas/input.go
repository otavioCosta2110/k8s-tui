package canvas

import (
	tea "github.com/charmbracelet/bubbletea"
)

func isKeyPressed(c Canvas, msg tea.KeyMsg) (tea.Model, tea.Cmd){
		if msg.String() == "q" {
			return &c, tea.Quit
		}
		if msg.String() == "a" {
      c.Height = c.Height/2
			return &c, nil
		}
    return &c, nil

}
