# Contributing

Create a terraform cli config file at `~/.terraformrc` with the following contents

Replace the path to this repo root path

```
provider_installation {
  dev_overrides {
    "Kamatera/kamatera" = "/path/to/go/src/github.com/Kamatera/terraform-provider-kamatera"
  }

  direct {}
}
```

Enable logging:

```
export TF_LOG_PROVIDER_KAMATERA=trace
```

* Install [Golang](https://golang.org/)
* Clone repo
* Build: `make build`
* Run terraform
* Update docs: `go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs`
* Commit, Push, Publish release
