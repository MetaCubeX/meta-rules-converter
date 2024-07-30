package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"convert/output/sing"

	"github.com/sagernet/sing-box/option"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	commandClash.PersistentFlags().StringVarP(&inPath, "in", "i", "", "Input Path")
	commandClash.PersistentFlags().StringVarP(&outType, "type", "t", "", "Output Type")
	commandClash.PersistentFlags().StringVarP(&outDir, "out", "o", "", "Output Path")
	mainCommand.AddCommand(commandClash)
}

var commandClash = &cobra.Command{
	Use:  "clash",
	RunE: convertClash,
}

func convertClash(cmd *cobra.Command, args []string) error {
	if inPath == "" {
		inPath = "."
	}
	if outDir == "" {
		outDir = "rules"
	}
	if outType == "" {
		outType = "sing-box"
	}
	outDir = strings.TrimSuffix(outDir, "/")
	files, err := searchFiles(inPath, ".yaml")
	if err != nil {
		return err
	}
	os.MkdirAll(outDir, 0755)

	var (
		domainFull    = make(map[string][]string)
		domainSuffix  = make(map[string][]string)
		domainKeyword = make(map[string][]string)
		domainRegex   = make(map[string][]string)
		ipCIDR        = make(map[string][]string)
		processName   = make(map[string][]string)
		packageName   = make(map[string][]string)
		processPath   = make(map[string][]string)
		dstPort       = make(map[string][]uint16)
	)

	for _, path := range files {
		var (
			d  []string
			ds []string
			dk []string
			dr []string
			ic []string
			pn []string
			pa []string
			pp []string
			dp []uint16
		)
		dirName, content := readFile(path)
		for _, line := range content {
			rule := strings.Split(line, ",")
			switch rule[0] {
			case "DOMAIN":
				d = append(d, rule[1])
			case "DOMAIN-SUFFIX":
				ds = append(ds, rule[1])
			case "DOMAIN-KEYWORD":
				dk = append(dk, rule[1])
			case "DOMAIN-REGEX":
				dr = append(dr, rule[1])
			case "IP-CIDR", "IP-CIDR6":
				ic = append(ic, rule[1])
			case "PROCESS-NAME":
				if strings.Contains(rule[1], ".exe") {
					pn = append(pn, rule[1])
				} else if strings.Contains(rule[1], ".") {
					pa = append(pa, rule[1])
				} else {
					pn = append(pn, rule[1])
				}
			case "PROCESS-PATH":
				pp = append(pp, rule[1])
			case "DST-PORT":
				port, _ := strconv.Atoi(rule[1])
				dp = append(dp, uint16(port))
			}
		}
		domainFull[dirName] = d
		domainSuffix[dirName] = ds
		domainKeyword[dirName] = dk
		domainRegex[dirName] = dr
		ipCIDR[dirName] = ic
		processName[dirName] = pn
		packageName[dirName] = pa
		processPath[dirName] = pp
		dstPort[dirName] = dp
	}

	switch outType {
	case "sing-box":
		for name, domain := range domainFull {
			os.MkdirAll(outDir+"/"+name, 0755)
			if len(domain) != 0 || len(domainSuffix[name]) != 0 || len(domainKeyword[name]) != 0 || len(domainRegex[name]) != 0 {
				domainRule := []option.DefaultHeadlessRule{
					{
						Domain:        domain,
						DomainKeyword: domainKeyword[name],
						DomainSuffix:  domainSuffix[name],
						DomainRegex:   domainRegex[name],
					},
				}
				sing.SaveSingRuleSet(domainRule, outDir+"/"+name+"/domain")
			}
			if len(ipCIDR[name]) != 0 {
				ipRule := []option.DefaultHeadlessRule{
					{
						IPCIDR: ipCIDR[name],
					},
				}
				sing.SaveSingRuleSet(ipRule, outDir+"/"+name+"/ip")
			}
			if len(processName[name]) != 0 || len(packageName[name]) != 0 || len(processPath[name]) != 0 {
				processRule := []option.DefaultHeadlessRule{
					{
						ProcessName: processName[name],
						PackageName: packageName[name],
						ProcessPath: processPath[name],
					},
				}
				sing.SaveSingRuleSet(processRule, outDir+"/"+name+"/process")
			}
			if len(dstPort[name]) != 0 {
				otherRule := []option.DefaultHeadlessRule{
					{
						Port: dstPort[name],
					},
				}
				sing.SaveSingRuleSet(otherRule, outDir+"/"+name+"/other")
			}
			classicalRule := []option.DefaultHeadlessRule{
				{
					Domain:        domainFull[name],
					DomainKeyword: domainKeyword[name],
					DomainSuffix:  domainSuffix[name],
					DomainRegex:   domainRegex[name],
					IPCIDR:        ipCIDR[name],
					ProcessName:   processName[name],
					PackageName:   packageName[name],
					ProcessPath:   processPath[name],
					Port:          dstPort[name],
				},
			}
			sing.SaveSingRuleSet(classicalRule, outDir+"/"+name+"/classical")
		}
	}

	return nil
}

func searchFiles(root string, keyword string) ([]string, error) {
	var result []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(info.Name(), keyword) {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}

func readFile(path string) (string, []string) {
	var dirName string
	if runtime.GOOS == "windows" {
		lastIndex := strings.LastIndex(path, "\\")
		if lastIndex != -1 {
			dirName = strings.ToLower(strings.TrimSuffix(path[lastIndex+1:], ".yaml"))
		}
	} else {
		lastIndex := strings.LastIndex(path, "/")
		if lastIndex != -1 {
			dirName = strings.ToLower(strings.TrimSuffix(path[lastIndex+1:], ".yaml"))
		}
	}
	file, _ := os.Open(path)
	data, _ := io.ReadAll(file)

	var rules Rules
	yaml.Unmarshal(data, &rules)
	return dirName, rules.Payload
}

type Rules struct {
	Payload []string `yaml:"payload"`
}
