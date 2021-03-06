package testutils

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/protolator"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer/etcdraft"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/genesisconfig"
	"github.com/kfsoftware/hlf-operator/internal/github.com/hyperledger/fabric/sdkinternal/configtxgen/encoder"
	genesisconfig2 "github.com/kfsoftware/hlf-operator/internal/github.com/hyperledger/fabric/sdkinternal/configtxgen/genesisconfig"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

type ConsortiumMember struct {
	MSPID string
}

type Consortium struct {
	Name          string
	Organizations []*ConsortiumMember
}

func ordererCapabilities() map[string]bool {
	return map[string]bool{
		"V2_0": true,
	}
}

const (
	consortiumDefault = "SampleConsortium"
)

type PeerNode struct {
	Host string
	Port int
}
type PeerOrganization struct {
	RootCert    string
	TLSRootCert string
	MspID       string
	Peers       []PeerNode
}

func memberToOrg(member PeerOrganization) (*genesisconfig.Organization, error) {
	serverTlsCertPem := []byte(member.TLSRootCert)
	serverRootCert := []byte(member.RootCert)
	mspID := member.MspID
	mspDir, err := ioutil.TempDir("", fmt.Sprintf("mspdir_%s", mspID))
	if err != nil {
		return nil, err
	}
	tlsCaCertsPath := path.Join(mspDir, "tlscacerts")
	err = os.Mkdir(tlsCaCertsPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	caCertsPath := path.Join(mspDir, "cacerts")
	err = os.Mkdir(caCertsPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(path.Join(caCertsPath, "cacert.pem"), serverRootCert, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(path.Join(tlsCaCertsPath, "tlscacert.pem"), serverTlsCertPem, os.ModePerm)
	if err != nil {
		return nil, err
	}
	configNodeOU := `
NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: orderer
`
	err = ioutil.WriteFile(path.Join(mspDir, "config.yaml"), []byte(configNodeOU), os.ModePerm)
	if err != nil {
		return nil, err
	}
	log.Printf("MSPDIR=%s", mspDir)
	anchorPeers := []*genesisconfig.AnchorPeer{}

	for _, peer := range member.Peers {
		anchorPeers = append(anchorPeers, &genesisconfig.AnchorPeer{
			Host: peer.Host,
			Port: peer.Port,
		})
	}
	genesisOrg := &genesisconfig.Organization{
		Name:    mspID,
		ID:      mspID,
		MSPDir:  mspDir,
		MSPType: "bccsp",
		Policies: map[string]*genesisconfig.Policy{
			"Admins": {
				Type: "Signature",
				Rule: fmt.Sprintf("OR('%s.admin')", mspID),
			},
			"Readers": {
				Type: "Signature",
				//Rule: fmt.Sprintf("OR('%s.admin', '%s.peer', '%s.client')", mspID, mspID, mspID),
				Rule: fmt.Sprintf("OR('%s.member')", mspID),
			},
			"Writers": {
				Type: "Signature",
				Rule: fmt.Sprintf("OR('%s.member')", mspID),
			},
		},
		AnchorPeers:   anchorPeers,
		SkipAsForeign: false,
	}
	return genesisOrg, nil
}

type OrdererNode struct {
	TLSCert string
	Host    string
	Port    int
}
type OrdererOrganization struct {
	Nodes        []OrdererNode
	RootTLSCert  string
	RootSignCert string
	MspID        string
}

func memberToOrgUpdate(member PeerOrganization) (*genesisconfig2.Organization, error) {
	serverTlsCertPem := []byte(member.TLSRootCert)
	rootCertPem := []byte(member.RootCert)
	mspID := member.MspID
	mspDir, err := ioutil.TempDir("", fmt.Sprintf("mspdir_%s", mspID))
	if err != nil {
		return nil, err
	}
	tlsCaCertsPath := path.Join(mspDir, "tlscacerts")
	err = os.Mkdir(tlsCaCertsPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	caCertsPath := path.Join(mspDir, "cacerts")
	err = os.Mkdir(caCertsPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(path.Join(caCertsPath, "cacert.pem"), rootCertPem, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(path.Join(tlsCaCertsPath, "tlscacert.pem"), serverTlsCertPem, os.ModePerm)
	if err != nil {
		return nil, err
	}
	configNodeOU := `
NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: orderer
`
	err = ioutil.WriteFile(path.Join(mspDir, "config.yaml"), []byte(configNodeOU), os.ModePerm)
	if err != nil {
		return nil, err
	}
	log.Printf("MSPDIR=%s", mspDir)
	anchorPeers := []*genesisconfig2.AnchorPeer{}
	for _, peer := range member.Peers {
		anchorPeers = append(anchorPeers, &genesisconfig2.AnchorPeer{
			Host: peer.Host,
			Port: peer.Port,
		})
	}
	genesisOrg := &genesisconfig2.Organization{
		Name:    mspID,
		ID:      mspID,
		MSPDir:  mspDir,
		MSPType: "bccsp",
		Policies: map[string]*genesisconfig2.Policy{
			"Admins": {
				Type: "Signature",
				Rule: fmt.Sprintf("OR('%s.admin')", mspID),
			},
			"Endorsement": {
				Type: "Signature",
				Rule: fmt.Sprintf("OR('%s.peer')", mspID),
			},
			"Readers": {
				Type: "Signature",
				Rule: fmt.Sprintf("OR('%s.admin', '%s.peer', '%s.client')", mspID, mspID, mspID),
			},
			"Writers": {
				Type: "Signature",
				Rule: fmt.Sprintf("OR('%s.admin', '%s.client')", mspID, mspID),
			},
		},
		AnchorPeers:   anchorPeers,
		SkipAsForeign: false,
	}
	return genesisOrg, nil
}
func GetUpdatedConfig(channelConfig *cb.Config, peerOrgs []PeerOrganization) (*cb.Config, error) {
	modifiedConfig := &cb.Config{}
	modifiedConfigBytes, err := proto.Marshal(channelConfig)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(modifiedConfigBytes, modifiedConfig)
	if err != nil {
		return nil, err
	}

	consortiumGroups := map[string]*cb.ConfigGroup{}
	for _, member := range peerOrgs {
		mspID := member.MspID
		peerOrg, err := memberToOrgUpdate(member)
		if err != nil {
			return nil, err
		}
		configGroup, err := encoder.NewConsortiumOrgGroup(peerOrg)
		if err != nil {
			return nil, err
		}
		consortiumGroups[mspID] = configGroup
	}
	modifiedConfig.ChannelGroup.Groups["Consortiums"].Groups[consortiumDefault].Groups = consortiumGroups
	return modifiedConfig, nil
}
func GetProfileConfig(ordOrgs []OrdererOrganization) (*genesisconfig.Profile, error) {
	organizations := []*genesisconfig.Organization{}
	ordererOrganizations := []*genesisconfig.Organization{}
	consenters := []*etcdraft.Consenter{}
	ordererAddresses := []string{}
	log.Printf("Orderer organization: %d", len(ordOrgs))
	for _, org := range ordOrgs {
		serverRootTlsCertPem := []byte(org.RootTLSCert)
		serverRootSignCertPem := []byte(org.RootSignCert)
		for _, node := range org.Nodes {
			clientTlsCertPem := []byte(node.TLSCert)
			serverTlsCertPem := []byte(node.TLSCert)
			clientCertFile, err := ioutil.TempFile("", "clientcert")
			if err != nil {
				return nil, err
			}
			err = ioutil.WriteFile(clientCertFile.Name(), clientTlsCertPem, 644)
			if err != nil {
				return nil, err
			}

			serverCertFile, err := ioutil.TempFile("", "servercert")
			if err != nil {
				return nil, err
			}
			err = ioutil.WriteFile(serverCertFile.Name(), serverTlsCertPem, 644)
			if err != nil {
				return nil, err
			}
			ordererHost := node.Host
			ordererPort := node.Port
			consenters = append(consenters, &etcdraft.Consenter{
				Host:          ordererHost,
				Port:          uint32(ordererPort),
				ClientTlsCert: []byte(clientCertFile.Name()),
				ServerTlsCert: []byte(serverCertFile.Name()),
			})

			ordererAddresses = append(ordererAddresses, fmt.Sprintf("%s:%d", ordererHost, ordererPort))
		}
		ordererOrgMspDir, err := ioutil.TempDir("", fmt.Sprintf("mspdir_%s", org.MspID))
		if err != nil {
			return nil, err
		}
		tlsCaCertsPath := path.Join(ordererOrgMspDir, "tlscacerts")
		err = os.Mkdir(tlsCaCertsPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
		caCertsPath := path.Join(ordererOrgMspDir, "cacerts")
		err = os.Mkdir(caCertsPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(path.Join(caCertsPath, "cacert.pem"), serverRootSignCertPem, os.ModePerm)
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(path.Join(tlsCaCertsPath, "tlscacert.pem"), serverRootTlsCertPem, os.ModePerm)
		if err != nil {
			return nil, err
		}
		configNodeOU := `
NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: orderer
`
		err = ioutil.WriteFile(path.Join(ordererOrgMspDir, "config.yaml"), []byte(configNodeOU), os.ModePerm)
		if err != nil {
			return nil, err
		}
		log.Printf("MSPDIR=%s", ordererOrgMspDir)
		ordererOrganizations = append(ordererOrganizations, &genesisconfig.Organization{
			Name:    org.MspID,
			ID:      org.MspID,
			MSPDir:  ordererOrgMspDir,
			MSPType: "bccsp",
			Policies: map[string]*genesisconfig.Policy{
				"Readers": {
					Type: "Signature",
					Rule: fmt.Sprintf("OR('%s.member')", org.MspID),
				},
				"Writers": {
					Type: "Signature",
					Rule: fmt.Sprintf("OR('%s.member')", org.MspID),
				},
				"Admins": {
					Type: "Signature",
					Rule: fmt.Sprintf("OR('%s.admin')", org.MspID),
				},
			},
			AnchorPeers:   nil,
			SkipAsForeign: false,
		})
	}
	profileConfig := &genesisconfig.Profile{
		Consortium: consortiumDefault,
		Application: &genesisconfig.Application{
			Organizations: organizations,
			Capabilities:  ordererCapabilities(),
			Policies: map[string]*genesisconfig.Policy{
				"Admins": {
					Type: "ImplicitMeta",
					Rule: "MAJORITY Admins",
				},
				"Endorsement": {
					Type: "ImplicitMeta",
					Rule: "MAJORITY Endorsement",
				},
				"LifecycleEndorsement": {
					Type: "ImplicitMeta",
					Rule: "MAJORITY Endorsement",
				},
				"Readers": {
					Type: "ImplicitMeta",
					Rule: "ANY Readers",
				},
				"Writers": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
			},
			ACLs: nil,
		},
		Orderer: &genesisconfig.Orderer{
			OrdererType: "etcdraft",
			EtcdRaft: &etcdraft.ConfigMetadata{
				Consenters: consenters,
				Options: &etcdraft.Options{
					SnapshotIntervalSize: 19,
					TickInterval:         "500ms",
					ElectionTick:         10,
					HeartbeatTick:        1,
					MaxInflightBlocks:    5,
				},
			},
			Addresses:    ordererAddresses,
			BatchTimeout: 2 * time.Second,
			BatchSize: genesisconfig.BatchSize{
				MaxMessageCount:   500,
				AbsoluteMaxBytes:  10 * 1024 * 1024,
				PreferredMaxBytes: 2 * 1024 * 1024,
			},
			Organizations: ordererOrganizations,
			Policies: map[string]*genesisconfig.Policy{
				"Readers": {
					Type: "ImplicitMeta",
					Rule: "ANY Readers",
				},
				"Writers": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
				"Admins": {
					Type: "ImplicitMeta",
					Rule: "ANY Admins",
				},
				"BlockValidation": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
			},
			Capabilities: ordererCapabilities(),
		},
		Consortiums: map[string]*genesisconfig.Consortium{
			consortiumDefault: {
				Organizations: organizations,
			},
		},
		Capabilities: ordererCapabilities(),
		Policies: map[string]*genesisconfig.Policy{
			"Admins": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Admins",
			},
			"Readers": {
				Type: "ImplicitMeta",
				Rule: "ANY Readers",
			},
			"Writers": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
		},
	}
	return profileConfig, nil
}

func GetConfigEnvelopeBytes(configUpdate *cb.ConfigUpdate) ([]byte, error) {
	var buf bytes.Buffer
	if err := protolator.DeepMarshalJSON(&buf, configUpdate); err != nil {
		return nil, err
	}
	channelConfigBytes, err := proto.Marshal(configUpdate)
	if err != nil {
		return nil, err
	}
	configUpdateEnvelope := &cb.ConfigUpdateEnvelope{
		ConfigUpdate: channelConfigBytes,
		Signatures:   nil,
	}
	configUpdateEnvelopeBytes, err := proto.Marshal(configUpdateEnvelope)
	if err != nil {
		return nil, err
	}
	payload := &cb.Payload{
		Data: configUpdateEnvelopeBytes,
	}
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return nil, err
	}
	configEnvelope := &cb.Envelope{
		Payload: payloadBytes,
	}
	return proto.Marshal(configEnvelope)
}

func GetChannelProfileConfig(ordService OrdererOrganization, members []PeerOrganization) (*genesisconfig.Profile, error) {
	var organizations []*genesisconfig.Organization

	var ordererOrganizations []*genesisconfig.Organization
	var consenters []*etcdraft.Consenter
	var ordererAddresses []string

	for _, node := range ordService.Nodes {
		clientTlsCertPem := []byte(node.TLSCert)
		serverTlsCertPem := []byte(node.TLSCert)
		clientCertFile, err := ioutil.TempFile("", "clientcert")
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(clientCertFile.Name(), clientTlsCertPem, 644)
		if err != nil {
			return nil, err
		}

		serverCertFile, err := ioutil.TempFile("", "servercert")
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(serverCertFile.Name(), serverTlsCertPem, 644)
		if err != nil {
			return nil, err
		}
		ordererHost := node.Host
		ordererPort := node.Port
		consenters = append(consenters, &etcdraft.Consenter{
			Host:          ordererHost,
			Port:          uint32(ordererPort),
			ClientTlsCert: []byte(clientCertFile.Name()),
			ServerTlsCert: []byte(serverCertFile.Name()),
		})

		ordererAddresses = append(ordererAddresses, fmt.Sprintf("%s:%d", ordererHost, ordererPort))
	}
	ordererOrgMspDir, err := ioutil.TempDir("", fmt.Sprintf("mspdir_%s", ordService.MspID))
	if err != nil {
		return nil, err
	}
	tlsCaCertsPath := path.Join(ordererOrgMspDir, "tlscacerts")
	err = os.Mkdir(tlsCaCertsPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	caCertsPath := path.Join(ordererOrgMspDir, "cacerts")
	err = os.Mkdir(caCertsPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	orgAdmin := members[0]
	serverRootTlsCertPem := []byte(orgAdmin.RootCert)
	err = ioutil.WriteFile(path.Join(caCertsPath, "cacert.pem"), serverRootTlsCertPem, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(path.Join(tlsCaCertsPath, "tlscacert.pem"), serverRootTlsCertPem, os.ModePerm)
	if err != nil {
		return nil, err
	}
	configNodeOU := `
NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: orderer
`
	err = ioutil.WriteFile(path.Join(ordererOrgMspDir, "config.yaml"), []byte(configNodeOU), os.ModePerm)
	if err != nil {
		return nil, err
	}
	log.Printf("MSPDIR=%s", ordererOrgMspDir)
	ordererOrganizations = append(ordererOrganizations, &genesisconfig.Organization{
		Name:    ordService.MspID,
		ID:      ordService.MspID,
		MSPDir:  ordererOrgMspDir,
		MSPType: "bccsp",
		Policies: map[string]*genesisconfig.Policy{
			"Readers": {
				Type: "Signature",
				Rule: fmt.Sprintf("OR('%s.member')", ordService.MspID),
			},
			"Writers": {
				Type: "Signature",
				Rule: fmt.Sprintf("OR('%s.member')", ordService.MspID),
			},
			"Admins": {
				Type: "Signature",
				Rule: fmt.Sprintf("OR('%s.admin')", ordService.MspID),
			},
		},
		AnchorPeers:   nil,
		SkipAsForeign: false,
	})
	log.Printf("Members: %d", len(members))
	for _, member := range members {
		peerOrg, err := memberToOrg(member)
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, peerOrg)
	}
	profileConfig := &genesisconfig.Profile{
		Consortium: consortiumDefault,
		Application: &genesisconfig.Application{
			Organizations: organizations,
			Capabilities:  ordererCapabilities(),
			Policies: map[string]*genesisconfig.Policy{
				"Admins": {
					Type: "Signature",
					Rule: fmt.Sprintf("OR('%s.admin')", orgAdmin.MspID),
				},
				"Endorsement": {
					Type: "ImplicitMeta",
					Rule: "MAJORITY Endorsement",
				},
				"LifecycleEndorsement": {
					Type: "ImplicitMeta",
					Rule: "MAJORITY Endorsement",
				},
				"Readers": {
					Type: "ImplicitMeta",
					Rule: "ANY Readers",
				},
				"Writers": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
			},
			ACLs: nil,
		},
		Orderer: &genesisconfig.Orderer{
			OrdererType: "etcdraft",
			EtcdRaft: &etcdraft.ConfigMetadata{
				Consenters: consenters,
				Options: &etcdraft.Options{
					SnapshotIntervalSize: 19,
					TickInterval:         "500ms",
					ElectionTick:         10,
					HeartbeatTick:        1,
					MaxInflightBlocks:    5,
				},
			},
			Addresses:    ordererAddresses,
			BatchTimeout: 2 * time.Second,
			BatchSize: genesisconfig.BatchSize{
				MaxMessageCount:   500,
				AbsoluteMaxBytes:  10 * 1024 * 1024,
				PreferredMaxBytes: 2 * 1024 * 1024,
			},
			Organizations: ordererOrganizations,
			Policies: map[string]*genesisconfig.Policy{
				"Readers": {
					Type: "ImplicitMeta",
					Rule: "ANY Readers",
				},
				"Writers": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
				"Admins": {
					Type: "ImplicitMeta",
					Rule: "ANY Admins",
				},
				"BlockValidation": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
			},
			Capabilities: ordererCapabilities(),
		},
		Capabilities: ordererCapabilities(),
		Policies: map[string]*genesisconfig.Policy{
			"Admins": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Admins",
			},
			"Readers": {
				Type: "ImplicitMeta",
				Rule: "ANY Readers",
			},
			"Writers": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
		},
	}
	return profileConfig, nil
}
