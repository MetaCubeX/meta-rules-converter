package meta

import (
	"os"

	P "github.com/metacubex/mihomo/constant/provider"
	RP "github.com/metacubex/mihomo/rules/provider"
)

func SaveMetaRuleSet(buf []byte, b string, f string, outputPath string) error {
	behavior, err := P.ParseBehavior(b)
	if err != nil {
		return err
	}
	format, err := P.ParseRuleFormat(f)
	if err != nil {
		return err
	}
	targetFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	RP.ConvertToMrs(buf, behavior, format, targetFile)
	err = targetFile.Close()
	if err != nil {
		return err
	}
	return nil
}
