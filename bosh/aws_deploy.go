package bosh

import (
	"github.com/EngineerBetter/concourse-up/bosh/internal/aws"
	"github.com/EngineerBetter/concourse-up/bosh/internal/boshenv"
	"github.com/EngineerBetter/concourse-up/db"
)

// Deploy implements deploy for AWS client
func (client *AWSClient) Deploy(state, creds []byte, detach bool) (newState, newCreds []byte, err error) {
	bosh, err := boshenv.New(boshenv.DownloadBOSH())
	if err != nil {
		return state, creds, err
	}

	state, creds, err = client.createEnv(bosh, state, creds)
	if err != nil {
		return state, creds, err
	}

	if err = client.updateCloudConfig(bosh); err != nil {
		return state, creds, err
	}
	if err = client.uploadConcourseStemcell(bosh); err != nil {
		return state, creds, err
	}
	if err = client.createDefaultDatabases(); err != nil {
		return state, creds, err
	}

	creds, err = client.deployConcourse(creds, detach)
	if err != nil {
		return state, creds, err
	}

	return state, creds, err
}

func (client *AWSClient) createEnv(bosh *boshenv.BOSHCLI, state, creds []byte) (newState, newCreds []byte, err error) {
	tags, err := splitTags(client.config.Tags)
	if err != nil {
		return state, creds, err
	}
	tags["concourse-up-project"] = client.config.Project
	tags["concourse-up-component"] = "concourse"
	//TODO(px): pull up this so that we use aws.Store
	store := temporaryStore{
		"vars.yaml":  creds,
		"state.json": state,
	}

	boshUserAccessKeyID, err1 := client.metadata.Get("BoshUserAccessKeyID")
	if err1 != nil {
		return state, creds, err1
	}
	boshSecretAccessKey, err1 := client.metadata.Get("BoshSecretAccessKey")
	if err1 != nil {
		return state, creds, err1
	}
	publicSubnetID, err1 := client.metadata.Get("PublicSubnetID")
	if err1 != nil {
		return state, creds, err1
	}
	privateSubnetID, err1 := client.metadata.Get("PrivateSubnetID")
	if err1 != nil {
		return state, creds, err1
	}
	directorPublicIP, err1 := client.metadata.Get("DirectorPublicIP")
	if err1 != nil {
		return state, creds, err1
	}
	atcSecurityGroupID, err1 := client.metadata.Get("ATCSecurityGroupID")
	if err1 != nil {
		return state, creds, err1
	}
	vmSecurityGroupID, err1 := client.metadata.Get("VMsSecurityGroupID")
	if err1 != nil {
		return state, creds, err1
	}
	blobstoreBucket, err1 := client.metadata.Get("BlobstoreBucket")
	if err1 != nil {
		return state, creds, err1
	}
	boshDBAddress, err1 := client.metadata.Get("BoshDBAddress")
	if err1 != nil {
		return state, creds, err1
	}
	boshDbPort, err1 := client.metadata.Get("BoshDBPort")
	if err1 != nil {
		return state, creds, err1
	}
	blobstoreUserAccessKeyID, err1 := client.metadata.Get("BlobstoreUserAccessKeyID")
	if err1 != nil {
		return state, creds, err1
	}
	blobstoreSecretAccessKey, err1 := client.metadata.Get("BlobstoreSecretAccessKey")
	if err1 != nil {
		return state, creds, err1
	}
	directorKeyPair, err1 := client.metadata.Get("DirectorKeyPair")
	if err1 != nil {
		return state, creds, err1
	}
	directorSecurityGroup, err1 := client.metadata.Get("DirectorSecurityGroupID")
	if err1 != nil {
		return state, creds, err1
	}
	err1 = bosh.CreateEnv(store, aws.Environment{
		InternalCIDR:    "10.0.0.0/24",
		InternalGateway: "10.0.0.1",
		InternalIP:      "10.0.0.6",
		AccessKeyID:     boshUserAccessKeyID,
		SecretAccessKey: boshSecretAccessKey,
		Region:          client.config.Region,
		AZ:              client.config.AvailabilityZone,
		DefaultKeyName:  directorKeyPair,
		DefaultSecurityGroups: []string{
			directorSecurityGroup,
			vmSecurityGroupID,
		},
		PrivateKey:           client.config.PrivateKey,
		PublicSubnetID:       publicSubnetID,
		PrivateSubnetID:      privateSubnetID,
		ExternalIP:           directorPublicIP,
		ATCSecurityGroup:     atcSecurityGroupID,
		VMSecurityGroup:      vmSecurityGroupID,
		BlobstoreBucket:      blobstoreBucket,
		DBCACert:             db.RDSRootCert,
		DBHost:               boshDBAddress,
		DBName:               client.config.RDSDefaultDatabaseName,
		DBPassword:           client.config.RDSPassword,
		DBPort:               boshDbPort,
		DBUsername:           client.config.RDSUsername,
		S3AWSAccessKeyID:     blobstoreUserAccessKeyID,
		S3AWSSecretAccessKey: blobstoreSecretAccessKey,
		Spot:                 client.config.Spot,
		WorkerType:           client.config.WorkerType}, client.config.DirectorPassword, client.config.DirectorCert, client.config.DirectorKey, client.config.DirectorCACert, tags)
	if err1 != nil {
		return store["state.json"], store["vars.yaml"], err1
	}
	return store["state.json"], store["vars.yaml"], err
}

func (client *AWSClient) updateCloudConfig(bosh *boshenv.BOSHCLI) error {
	publicSubnetID, err := client.metadata.Get("PublicSubnetID")
	if err != nil {
		return err
	}
	privateSubnetID, err := client.metadata.Get("PrivateSubnetID")
	if err != nil {
		return err
	}
	aTCSecurityGroupID, err := client.metadata.Get("ATCSecurityGroupID")
	if err != nil {
		return err
	}
	vMsSecurityGroupID, err := client.metadata.Get("VMsSecurityGroupID")
	if err != nil {
		return err
	}
	directorPublicIP, err := client.metadata.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return bosh.UpdateCloudConfig(aws.Environment{
		AZ:               client.config.AvailabilityZone,
		PublicSubnetID:   publicSubnetID,
		PrivateSubnetID:  privateSubnetID,
		ATCSecurityGroup: aTCSecurityGroupID,
		VMSecurityGroup:  vMsSecurityGroupID,
		Spot:             client.config.Spot,
		ExternalIP:       directorPublicIP,
		WorkerType:       client.config.WorkerType,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)

	return nil
}
func (client *AWSClient) uploadConcourseStemcell(bosh *boshenv.BOSHCLI) error {
	directorPublicIP, err := client.metadata.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return bosh.UploadConcourseStemcell(aws.Environment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)
	return nil
}

func (client *AWSClient) saveStateFile(bytes []byte) (string, error) {
	if bytes == nil {
		return client.director.PathInWorkingDir(StateFilename), nil
	}

	return client.director.SaveFileToWorkingDir(StateFilename, bytes)
}

func (client *AWSClient) saveCredsFile(bytes []byte) (string, error) {
	if bytes == nil {
		return client.director.PathInWorkingDir(CredsFilename), nil
	}

	return client.director.SaveFileToWorkingDir(CredsFilename, bytes)
}
