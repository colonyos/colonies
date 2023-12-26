package table

import (
	"fmt"

	"github.com/muesli/termenv"
)

var (
	env = termenv.EnvColorProfile()
)

// Theme defines a color theme used for printing tables.
type Theme struct {
	ColorRed     termenv.Color
	ColorYellow  termenv.Color
	ColorGreen   termenv.Color
	ColorBlue    termenv.Color
	ColorGray    termenv.Color
	ColorMagenta termenv.Color
	ColorCyan    termenv.Color
}

func defaultThemeName() string {
	if !termenv.HasDarkBackground() {
		return "light"
	}
	return "dark"
}

func LoadTheme(theme string) (Theme, error) {
	themes := make(map[string]Theme)

	themes["dark"] = Theme{
		ColorRed:     env.Color("#E88388"),
		ColorYellow:  env.Color("#DBAB79"),
		ColorGreen:   env.Color("#A8CC8C"),
		ColorBlue:    env.Color("#71BEF2"),
		ColorGray:    env.Color("#B9BFCA"),
		ColorMagenta: env.Color("#D290E4"),
		ColorCyan:    env.Color("#66C2CD"),
	}

	themes["light"] = Theme{
		ColorRed:     env.Color("#D70000"),
		ColorYellow:  env.Color("#FFAF00"),
		ColorGreen:   env.Color("#005F00"),
		ColorBlue:    env.Color("#000087"),
		ColorGray:    env.Color("#303030"),
		ColorMagenta: env.Color("#AF00FF"),
		ColorCyan:    env.Color("#0087FF"),
	}

	themes["ansi"] = Theme{
		ColorRed:     env.Color("9"),
		ColorYellow:  env.Color("11"),
		ColorGreen:   env.Color("10"),
		ColorBlue:    env.Color("12"),
		ColorGray:    env.Color("7"),
		ColorMagenta: env.Color("13"),
		ColorCyan:    env.Color("8"),
	}

	if _, ok := themes[theme]; !ok {
		return Theme{}, fmt.Errorf("unknown theme: %s", theme)
	}

	return themes[theme], nil
}
