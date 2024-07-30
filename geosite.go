package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"convert/output/meta"
	"convert/output/sing"

	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"

	"github.com/sagernet/sing-box/option"
	"github.com/spf13/cobra"
	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
)

func init() {
	commandSite.PersistentFlags().StringVarP(&inPath, "file", "f", "", "Input File Path")
	commandSite.PersistentFlags().StringVarP(&outType, "type", "t", "", "Output Type")
	commandSite.PersistentFlags().StringVarP(&outDir, "out", "o", "", "Output Path")
	mainCommand.AddCommand(commandSite)
}

var commandSite = &cobra.Command{
	Use:  "geosite",
	RunE: convertSite,
}

func convertSite(cmd *cobra.Command, args []string) error {
	if inPath == "" {
		inPath = "geosite.dat"
	}
	if outDir == "" {
		outDir = "geosite"
	}
	if outType == "" {
		outType = "clash"
	}
	outDir = strings.TrimSuffix(outDir, "/")
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}
	os.MkdirAll(outDir, 0755)

	var (
		domains       = make(map[string][]string)
		classical     = make(map[string][]string)
		domainFull    = make(map[string][]string)
		domainSuffix  = make(map[string][]string)
		domainKeyword = make(map[string][]string)
		domainRegex   = make(map[string][]string)
		wg            sync.WaitGroup
		mutex         sync.Mutex
	)

	list := routercommon.GeoSiteList{}
	err = proto.Unmarshal(data, &list)
	if err != nil {
		return err
	}
	for _, entry := range list.Entry {
		wg.Add(1)
		go func(entry *routercommon.GeoSite) {
			defer wg.Done()
			code := strings.ToLower(entry.CountryCode)
			marks := make(map[string][]*routercommon.Domain)
			var (
				d []string
				c []string
				f []string
				s []string
				k []string
				r []string
			)
			for _, domain := range entry.Domain {
				if len(domain.Attribute) > 0 {
					for _, attribute := range domain.Attribute {
						marks[attribute.Key] = append(marks[attribute.Key], domain)
					}
				}
				switch domain.Type {
				case routercommon.Domain_Full:
					d = append(d, domain.Value)
					c = append(c, "DOMAIN,"+domain.Value)
					f = append(f, domain.Value)
				case routercommon.Domain_RootDomain:
					d = append(d, "+."+domain.Value)
					c = append(c, "DOMAIN-SUFFIX,"+domain.Value)
					s = append(s, domain.Value)
				case routercommon.Domain_Regex:
					c = append(c, "DOMAIN-REGEX,"+domain.Value)
					r = append(r, domain.Value)
				case routercommon.Domain_Plain:
					c = append(c, "DOMAIN-KEYWORD,"+domain.Value)
					k = append(k, domain.Value)
				}
			}
			mutex.Lock()
			switch outType {
			case "clash":
				domains[code] = d
				classical[code] = c
			case "sing-box":
				domainFull[code] = f
				domainSuffix[code] = s
				domainKeyword[code] = k
				domainRegex[code] = r
			}
			mutex.Unlock()

			for mark, markEntries := range marks {
				var (
					md []string
					mc []string
					mf []string
					ms []string
					mk []string
					mr []string
				)
				for _, domain := range markEntries {
					switch domain.Type {
					case routercommon.Domain_Full:
						md = append(md, domain.Value)
						mc = append(mc, "DOMAIN,"+domain.Value)
						mf = append(mf, domain.Value)
					case routercommon.Domain_RootDomain:
						md = append(md, "+."+domain.Value)
						mc = append(mc, "DOMAIN-SUFFIX,"+domain.Value)
						ms = append(ms, domain.Value)
					case routercommon.Domain_Regex:
						mc = append(mc, "DOMAIN-REGEX,"+domain.Value)
						mr = append(mr, domain.Value)
					case routercommon.Domain_Plain:
						mc = append(mc, "DOMAIN-KEYWORD,"+domain.Value)
						mk = append(mk, domain.Value)
					}
				}
				mutex.Lock()
				switch outType {
				case "clash":
					domains[code+"@"+mark] = md
					classical[code+"@"+mark] = mc
				case "sing-box":
					domainFull[code+"@"+mark] = mf
					domainSuffix[code+"@"+mark] = ms
					domainKeyword[code+"@"+mark] = mk
					domainRegex[code+"@"+mark] = mr
				}
				mutex.Unlock()
			}
		}(entry)
	}
	wg.Wait()

	switch outType {
	case "clash":
		os.MkdirAll(outDir+"/classical", 0755)
		for code, domain := range domains {
			domainMap := map[string][]string{
				"payload": domain,
			}
			domainOut, err := yaml.Marshal(&domainMap)
			if err != nil {
				fmt.Println(code, " coding err: ", err)
			}
			err = os.WriteFile(outDir+"/"+code+".yaml", domainOut, 0755)
			if err != nil {
				fmt.Println(code, " output err: ", err)
			}
			err = os.WriteFile(outDir+"/"+code+".list", []byte(strings.Join(domain, "\n")), 0755)
			if err != nil {
				fmt.Println(code, " output err: ", err)
			}
			err = meta.SaveMetaRuleSet(domainOut, "domain", "yaml", outDir+"/"+code+".mrs")
			if err != nil {
				fmt.Println(code, " output err: ", err)
			}
			classicalMap := map[string][]string{
				"payload": classical[code],
			}
			classicalOut, err := yaml.Marshal(&classicalMap)
			if err != nil {
				fmt.Println(code, " coding err: ", err)
			}
			err = os.WriteFile(outDir+"/classical/"+code+".yaml", classicalOut, 0755)
			if err != nil {
				fmt.Println(code, " output err: ", err)
			}
			err = os.WriteFile(outDir+"/classical/"+code+".list", []byte(strings.Join(classical[code], "\n")), 0755)
			if err != nil {
				fmt.Println(code, " output err: ", err)
			}
			// meta.SaveMetaRuleSet(classicalOut, "classical", "yaml", outDir+"/classical/"+code+".mrs")
		}
	case "sing-box":
		for code, domain := range domainFull {
			domainRule := []option.DefaultHeadlessRule{
				{
					Domain:        domain,
					DomainKeyword: domainKeyword[code],
					DomainSuffix:  domainSuffix[code],
					DomainRegex:   domainRegex[code],
				},
			}
			sing.SaveSingRuleSet(domainRule, outDir+"/"+code)
		}
	}
	return nil
}
