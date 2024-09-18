package main

/*
 * This is a tool that can be used to see how Domain Connect template needs
 * and/or can be used in 'http...apply?' request.
 *
 * Questions about the tool can be sent to <domain-connect@cloudflare.com>
 */

import (
	"flag"
	"fmt"
	"os"
)

type parameters struct {
	host           *string
	redirect_url   *string
	state          *string
	kvs            *string
	providerName   *string
	serviceName    *string
	groupId        *string
	privateKeyPath *string
	key            *string
}

func main() {
	var params parameters

	// Command line option handling
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <./template.json> <example.com>\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "See also https://github.com/Domain-Connect/spec/blob/master/Domain%%20Connect%%20Spec%%20Draft.adoc\n")
	}
	params.host = flag.String("param.host", "", "host query parameter")
	params.redirect_url = flag.String("param.redirect_url", "", "redirect_url query parameter")
	params.state = flag.String("param.state", "", "state query parameter")
	params.kvs = flag.String("param.kvs", "", "kvs query parameters, for example: '%key1%val1%key2%val2%'")
	params.providerName = flag.String("param.providerName", "", "providerName query parameter")
	params.serviceName = flag.String("param.serviceName", "", "serviceName query parameter")
	params.groupId = flag.String("param.groupId", "", "groupId query parameter")
	params.key = flag.String("param.key", "", "key query parameter")
	params.privateKeyPath = flag.String("cmd.privatekey", "", "path to private key file, this generates 'sig', see https://exampleservice.domainconnect.org/sig")
	performVersionComparison := flag.Bool("cmd.checkversions", true, "compare template file and service provider version")
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Println("not enough arguments, try --help")
		os.Exit(1)
	}

	template, err := readTemplate(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	settings, err := getSettings(flag.Arg(1))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *performVersionComparison {
		if ok, err := compareVersions(template, settings); err != nil || !ok {
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("version mismatch")
			}
			os.Exit(1)
		}
	}

	kvMap := fromCmdline(*params.kvs)
	kvMap = crossReference(template, kvMap, *params.groupId)

	printApply(flag.Arg(1), template, settings, &params, kvMap)
}
