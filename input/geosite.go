package input

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/metacubex/meta-rules-converter/output/meta"
	"github.com/metacubex/meta-rules-converter/output/sing"

	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"

	"github.com/metacubex/mihomo/component/geodata/router"
	"github.com/spf13/cobra"
)

func ConvertSite(cmd *cobra.Command, inPath string, outType string, outDir string) error {
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
	os.MkdirAll(outDir, 0777)

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

	list := router.GeoSiteList{}
	err = proto.Unmarshal(data, &list)
	if err != nil {
		return err
	}
	for _, entry := range list.Entry {
		wg.Add(1)
		go func(entry *router.GeoSite) {
			defer wg.Done()
			code := strings.ToLower(entry.CountryCode)
			marks := make(map[string][]*router.Domain)
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
				case router.Domain_Full:
					d = append(d, domain.Value)
					c = append(c, "DOMAIN,"+domain.Value)
					f = append(f, domain.Value)
				case router.Domain_Domain:
					d = append(d, "+."+domain.Value)
					c = append(c, "DOMAIN-SUFFIX,"+domain.Value)
					s = append(s, domain.Value)
				case router.Domain_Regex:
					c = append(c, "DOMAIN-REGEX,"+domain.Value)
					r = append(r, domain.Value)
				case router.Domain_Plain:
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
					case router.Domain_Full:
						md = append(md, domain.Value)
						mc = append(mc, "DOMAIN,"+domain.Value)
						mf = append(mf, domain.Value)
					case router.Domain_Domain:
						md = append(md, "+."+domain.Value)
						mc = append(mc, "DOMAIN-SUFFIX,"+domain.Value)
						ms = append(ms, domain.Value)
					case router.Domain_Regex:
						mc = append(mc, "DOMAIN-REGEX,"+domain.Value)
						mr = append(mr, domain.Value)
					case router.Domain_Plain:
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
		os.MkdirAll(outDir+"/classical", 0777)
		for code, domain := range domains {
			domainMap := map[string][]string{
				"payload": domain,
			}
			domainOut, err := yaml.Marshal(&domainMap)
			if err != nil {
				fmt.Println(code, " coding err: ", err)
			}
			err = os.WriteFile(outDir+"/"+code+".yaml", domainOut, 0666)
			if err != nil {
				fmt.Println(code, " output err: ", err)
			}
			domainOut = []byte(strings.Join(domain, "\n"))
			err = os.WriteFile(outDir+"/"+code+".list", domainOut, 0666)
			if err != nil {
				fmt.Println(code, " output err: ", err)
			}
			err = meta.SaveMetaRuleSet(domainOut, "domain", "text", outDir+"/"+code+".mrs")
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
			err = os.WriteFile(outDir+"/classical/"+code+".yaml", classicalOut, 0666)
			if err != nil {
				fmt.Println(code, " output err: ", err)
			}
			classicalOut = []byte(strings.Join(classical[code], "\n"))
			err = os.WriteFile(outDir+"/classical/"+code+".list", classicalOut, 0666)
			if err != nil {
				fmt.Println(code, " output err: ", err)
			}
			// meta.SaveMetaRuleSet(classicalOut, "classical", "text", outDir+"/classical/"+code+".mrs")
		}
	case "sing-box":
		for code, domain := range domainFull {
			domainRule := []sing.DefaultHeadlessRule{
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
