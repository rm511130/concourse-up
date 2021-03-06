package concourse

import (
	"io"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/commands/deploy"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/director"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
)

// Client is a concrete implementation of IClient interface
type Client struct {
	provider              iaas.Provider
	tfCLI                 terraform.CLIInterface
	boshClientFactory     bosh.ClientFactory
	flyClientFactory      func(fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error)
	certGenerator         func(constructor func(u *certs.User) (certs.AcmeClient, error), caName string, ip ...string) (*certs.Certs, error)
	configClient          config.IClient
	deployArgs            *deploy.Args
	stdout                io.Writer
	stderr                io.Writer
	ipChecker             func() (string, error)
	acmeClientConstructor func(u *certs.User) (certs.AcmeClient, error)
	versionFile           []byte
	version               string
}

// IClient represents a concourse-up client
type IClient interface {
	Deploy() error
	Destroy() error
	FetchInfo() (*Info, error)
}

//go:generate go-bindata -pkg $GOPACKAGE ../../concourse-up-ops/director-versions.json
var versionFile = MustAsset("../../concourse-up-ops/director-versions.json")

// NewClient returns a new Client
func NewClient(
	provider iaas.Provider,
	tfCLI terraform.CLIInterface,
	boshClientFactory bosh.ClientFactory,
	flyClientFactory func(fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error),
	certGenerator func(constructor func(u *certs.User) (certs.AcmeClient, error), caName string, ip ...string) (*certs.Certs, error),
	configClient config.IClient,
	deployArgs *deploy.Args,
	stdout, stderr io.Writer,
	ipChecker func() (string, error),
	acmeClientConstructor func(u *certs.User) (certs.AcmeClient, error),
	version string) *Client {
	return &Client{
		provider:              provider,
		tfCLI:                 tfCLI,
		boshClientFactory:     boshClientFactory,
		flyClientFactory:      flyClientFactory,
		configClient:          configClient,
		certGenerator:         certGenerator,
		deployArgs:            deployArgs,
		stdout:                stdout,
		stderr:                stderr,
		ipChecker:             ipChecker,
		acmeClientConstructor: acmeClientConstructor,
		versionFile:           versionFile,
		version:               version,
	}
}

func (client *Client) buildBoshClient(config config.Config, metadata terraform.IAASMetadata) (bosh.IClient, error) {
	directorPublicIP, err := metadata.Get("DirectorPublicIP")
	if err != nil {
		return nil, err
	}
	director, err := director.NewClient(director.Credentials{
		Username: config.DirectorUsername,
		Password: config.DirectorPassword,
		Host:     directorPublicIP,
		CACert:   config.DirectorCACert,
	}, versionFile)
	if err != nil {
		return nil, err
	}

	return client.boshClientFactory(
		config,
		metadata,
		director,
		client.stdout,
		client.stderr,
		client.provider,
	)
}
