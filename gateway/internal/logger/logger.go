package logger

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/NikitaYamukov/go-microservices/gateway/internal/config"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

func New(cfg *config.Config) zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatLevel: func(i interface{}) string {
			if i == nil {
				return ""
			}
			level := strings.ToUpper(fmt.Sprintf("%s", i))
			switch level {
			case "INFO":
				return color.New(color.FgGreen).Sprint("INF")
			case "WARN":
				return color.New(color.FgYellow).Sprint("WRN")
			case "ERROR":
				return color.New(color.FgRed).Sprint("ERR")
			case "DEBUG":
				return color.New(color.FgBlue).Sprint("DBG")
			default:
				return level
			}
		},
	}

	log := zerolog.New(writer).
		With().
		Timestamp().
		Str("source", cfg.ServiceName).
		Logger()

	// Настройки spew
	spew.Config.Indent = "  "
	spew.Config.DisablePointerAddresses = true
	spew.Config.DisableCapacities = true
	spew.Config.SortKeys = true

	// Генерируем дамп
	dump := spew.Sdump(cfg)

	// Регэксп для поиска ключей struct'ов в формате spew:  "FieldName:"
	keyColor := color.New(color.FgCyan).SprintFunc()
	re := regexp.MustCompile(`(?m)^(\s*)([A-Za-z0-9_]+):`)
	coloredDump := re.ReplaceAllStringFunc(dump, func(s string) string {
		matches := re.FindStringSubmatch(s)
		if len(matches) != 3 {
			return s
		}
		return fmt.Sprintf("%s%s:", matches[1], keyColor(matches[2]))
	})

	log.Info().Msg("Loaded configuration:")
	log.Info().Msg(coloredDump)

	return log
}
