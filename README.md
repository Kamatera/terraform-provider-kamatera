# Terraform Provider for Kamatera

## Usage

Set environment variables

```
export KAMATERA_API_CLIENT_ID=
export KAMATERA_API_SECRET=
```

Apply

```
terraform apply tests
```


## Development

* Install [Golang](https://golang.org/)
* Clone repo
* `go build -o terraform-provider-kamatera`
* `terraform init tests`
