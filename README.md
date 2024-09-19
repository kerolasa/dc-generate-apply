## Domain Connect generate template apply url

This tool demonstrates to Service Providers how to generate a
[Domain Connect](https://github.com/Domain-Connect/spec/blob/master/Domain%20Connect%20Spec%20Draft.adoc)
apply url targetting a DNS Provider.

Notice that you need a private key to generate signature.  See
[signature help page](https://exampleservice.domainconnect.org/sig)
how to generate a private key.

```
go install github.com/kerolasa/dc-generate-apply@latest
$GOPATH/dc-generate-apply --help
Usage: ./dc-generate-apply [options] -cmd.privatekey ./key.pem ./template.json example.com
  -cmd.checkversions
	compare template file and service provider version (default true)
  -cmd.privatekey string
	path to private key file, this generates 'sig', see https://exampleservice.domainconnect.org/sig
  -loglevel string
	loglevel can be one of: panic fatal error warn info debug trace (default "info")
  -param.groupId string
	groupId query parameter
  -param.host string
	host query parameter
  -param.key string
	key query parameter
  -param.kvs string
	kvs query parameters, for example: '%key1%val1%key2%val2%'
  -param.providerName string
	providerName query parameter
  -param.redirect_url string
	redirect_url query parameter
  -param.serviceName string
	serviceName query parameter
  -param.state string
	state query parameter
See also https://github.com/Domain-Connect/spec/blob/master/Domain%20Connect%20Spec%20Draft.adoc
```
