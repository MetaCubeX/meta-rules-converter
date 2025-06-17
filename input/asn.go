package input

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/metacubex/meta-rules-converter/output/meta"
	"github.com/metacubex/meta-rules-converter/output/sing"
	"github.com/oschwald/maxminddb-golang"
	"github.com/spf13/cobra"
)

type ASN struct {
	AutonomousSystemNumber       uint   `maxminddb:"autonomous_system_number"`
	AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
}

func ConvertASN(cmd *cobra.Command, inPath string, outType string, outDir string) error {
	if inPath == "" {
		inPath = "GeoLite2-ASN.mmdb"
	}
	if outDir == "" {
		outDir = "asn"
	}
	if outType == "" {
		outType = "clash"
	}
	outDir = strings.TrimSuffix(outDir, "/")
	db, err := maxminddb.Open(inPath)
	if err != nil {
		return err
	}
	defer db.Close()

	os.MkdirAll(outDir, 0777)

	countryCIDRs := make(map[uint][]string)
	networks := db.Networks(maxminddb.SkipAliasedNetworks)

	for networks.Next() {
		var asn ASN
		network, err := networks.Network(&asn)
		if err != nil {
			fmt.Printf("Error decoding network: %v", err)
			continue
		}
		countryCIDRs[asn.AutonomousSystemNumber] = append(countryCIDRs[asn.AutonomousSystemNumber], network.String())
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 100)

	switch outType {
	case "clash":
		for number, cidrs := range countryCIDRs {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(number uint, cidrs []string) {
				defer wg.Done()
				defer func() { <-semaphore }()
				code := fmt.Sprintf("AS%d", number)

				err := os.WriteFile(outDir+"/"+code+".list", []byte(strings.Join(cidrs, "\n")), 0666)
				if err != nil {
					fmt.Println(code, " output err: ", err)
				}
				err = meta.SaveMetaRuleSet([]byte(strings.Join(cidrs, "\n")), "ipcidr", "text", filepath.Join(outDir, code+".mrs"))
				if err != nil {
					fmt.Printf("%s output err: %v", code, err)
				}
			}(number, cidrs)
		}
	case "sing-box":
		for number, cidrs := range countryCIDRs {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(number uint, cidrs []string) {
				defer wg.Done()
				defer func() { <-semaphore }()
				ipcidrRule := []sing.DefaultHeadlessRule{{IPCIDR: cidrs}}
				err := sing.SaveSingRuleSet(ipcidrRule, filepath.Join(outDir, fmt.Sprintf("AS%d", number)))
				if err != nil {
					fmt.Printf("AS%d output err: %v", number, err)
				}
			}(number, cidrs)
		}
	}

	wg.Wait()
	return nil
}
