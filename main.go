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
	"strconv"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	if isatty.IsTerminal(os.Stderr.Fd()) {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.RFC3339,
			},
		)
	}
	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
		return file + ":" + strconv.Itoa(line)
	}
	log.Logger = log.With().Caller().Logger()

	var params parameters

	// Command line option handling
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] -cmd.privatekey ./key.pem ./template.json example.com\n", os.Args[0])
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
	loglevel := flag.String("loglevel", "info", "loglevel can be one of: panic fatal error warn info debug trace")
	flag.Parse()

	level, err := zerolog.ParseLevel(*loglevel)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid loglevel")
	}
	zerolog.SetGlobalLevel(level)

	if flag.NArg() < 2 {
		log.Fatal().Msg("not enough arguments, try --help")
	}

	template := readTemplate(flag.Arg(0))

	settings := getSettings(flag.Arg(1))

	if *performVersionComparison {
		compareVersions(template, settings)
	}

	kvMap := fromCmdline(*params.kvs)
	kvMap = crossReference(template, kvMap, *params.groupId)

	printApply(flag.Arg(1), template, settings, &params, kvMap)
}
