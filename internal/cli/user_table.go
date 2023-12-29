package cli

import (
	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
)

func printUserTable(user *core.User) {
	t, theme := createTable(1)

	row := []interface{}{
		termenv.String("Username").Foreground(theme.ColorCyan),
		termenv.String(user.Name).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("UserId").Foreground(theme.ColorCyan),
		termenv.String(user.ID).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Colony").Foreground(theme.ColorCyan),
		termenv.String(user.ColonyName).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Email").Foreground(theme.ColorCyan),
		termenv.String(user.Email).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	row = []interface{}{
		termenv.String("Phone").Foreground(theme.ColorCyan),
		termenv.String(user.Phone).Foreground(theme.ColorGray),
	}
	t.AddRow(row)

	t.Render()
}

func printUsersTable(users []*core.User) {
	t, theme := createTable(1)

	var cols = []table.Column{
		{ID: "Username", Name: "Username", SortIndex: 1},
		{ID: "Email", Name: "email", SortIndex: 2},
		{ID: "Phone", Name: "Phone", SortIndex: 3},
	}
	t.SetCols(cols)

	for _, user := range users {
		row := []interface{}{
			termenv.String(user.Name).Foreground(theme.ColorCyan),
			termenv.String(user.Email).Foreground(theme.ColorViolet),
			termenv.String(user.Phone).Foreground(theme.ColorMagenta),
		}
		t.AddRow(row)
	}

	t.Render()
}
