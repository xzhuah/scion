// Copyright 2018 ETH Zurich, Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/scionproto/scion/go/cert_srv/internal/config"
	"github.com/scionproto/scion/go/cert_srv/internal/drkey"
	"github.com/scionproto/scion/go/cert_srv/internal/metrics"
	"github.com/scionproto/scion/go/cert_srv/internal/reiss"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/drkeystorage"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/infra/modules/trust"
	"github.com/scionproto/scion/go/lib/infra/modules/trust/trustdb"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/topology"
	"github.com/scionproto/scion/go/proto"
)

const (
	ErrorConf      = "Unable to load configuration"
	ErrorDispClose = "Unable to close dispatcher"
	ErrorDispInit  = "Unable to initialize dispatcher"
	ErrorSign      = "Unable to create sign"
	ErrorSNET      = "Unable to create local SCION Network context"
)

// setupBasic loads the config from file and initializes logging.
func setupBasic() error {
	// Load and initialize config.
	if _, err := toml.DecodeFile(env.ConfigFile(), &cfg); err != nil {
		return err
	}
	cfg.InitDefaults()
	if err := env.InitLogging(&cfg.Logging); err != nil {
		return err
	}
	metrics.Init(cfg.General.ID)
	return env.LogAppStarted(common.CS, cfg.General.ID)
}

// setup initializes the config and sets the messenger.
func setup() error {
	if err := cfg.Validate(); err != nil {
		return common.NewBasicError("Unable to validate config", err)
	}
	itopo.Init(cfg.General.ID, proto.ServiceType_cs, itopo.Callbacks{})
	topo, err := topology.LoadFromFile(cfg.General.Topology)
	if err != nil {
		return common.NewBasicError("Unable to load topology", err)
	}
	if _, _, err := itopo.SetStatic(topo, false); err != nil {
		return common.NewBasicError("Unable to set initial static topology", err)
	}
	// Set environment to listen for signals.
	infraenv.InitInfraEnvironmentFunc(cfg.General.Topology, func() {
		if err := reload(); err != nil {
			log.Error("Unable to reload", "err", err)
		}
	})
	router, err := infraenv.NewRouter(topo.ISD_AS, cfg.Sciond)
	if err != nil {
		return common.NewBasicError("Unable to initialize path router", err)
	}
	// Load CS state.
	if err := initState(&cfg, router); err != nil {
		return common.NewBasicError("Unable to initialize CS state", err)
	}
	if err := setMessenger(&cfg, router); err != nil {
		return common.NewBasicError("Unable to set messenger", err)
	}
	if err := addHandlers(); err != nil {
		return common.NewBasicError("Unable to add handlers", err)
	}
	return nil
}

// reload reloads the topology and CS config.
func reload() error {
	// FIXME(roosd): KeyConf reloading is not yet supported.
	// https://github.com/scionproto/scion/issues/2077
	var newConf config.Config
	// Load new config to get the CS parameters.
	if _, err := toml.DecodeFile(env.ConfigFile(), &newConf); err != nil {
		return err
	}
	newConf.InitDefaults()
	if err := newConf.Validate(); err != nil {
		return common.NewBasicError("Unable to validate new config", err)
	}
	cfg.CS = newConf.CS
	// Restart the periodic reissue task to respect the fresh parameters.
	stopReissRunner()
	startReissRunner()
	stopDRKeyRunners()
	startDRKeyRunners()
	return nil
}

// initState sets the state.
func initState(cfg *config.Config, router snet.Router) error {
	topo := itopo.Get()
	var err error
	if trustDB, err = cfg.TrustDB.New(); err != nil {
		return common.NewBasicError("Unable to initialize trustDB", err)
	}
	trustDB = trustdb.WithMetrics("std", trustDB)
	trustConf := trust.Config{
		MustHaveLocalChain: true,
		ServiceType:        proto.ServiceType_cs,
		Router:             router,
		TopoProvider:       itopo.Provider(),
	}
	trustStore := trust.NewStore(trustDB, topo.ISD_AS, trustConf, log.Root())
	err = trustStore.LoadAuthoritativeCrypto(filepath.Join(cfg.General.ConfigDir, "certs"))
	if err != nil {
		return common.NewBasicError("Unable to load local crypto", err)
	}
	keyConf, err := config.LoadKeyConf(cfg.General.ConfigDir, topo.Core)
	if err != nil {
		return common.NewBasicError("Unable to load AS keys", err)
	}
	var drkeyStore drkeystorage.ServiceStore
	var svFactory drkeystorage.SecretValueFactory
	if cfg.DRKey.Enabled() {
		svFactory = drkey.NewSecretValueFactory(
			keyConf.Master.Key0, cfg.DRKey.EpochDuration.Duration)
		drkeyDB, err := cfg.DRKey.NewDB()
		if err != nil {
			return common.NewBasicError("Unable to initialize DRKey DB", err)
		}
		drkeyStore = drkey.NewServiceStore(topo.ISD_AS, keyConf.DecryptKey,
			drkeyDB, trustDB, svFactory, cfg.DRKey.Delegation.ToMapPerHost())
		log.Info("DRKey is enabled")
	} else {
		log.Warn("DRKey is DISABLED by configuration")
	}
	state = config.NewState(keyConf, trustDB, trustStore, svFactory, drkeyStore)
	if err = setDefaultSignerVerifier(state, topo.ISD_AS); err != nil {
		return common.NewBasicError("Unable to set default signer and verifier", err)
	}
	return nil
}

// setDefaultSignerVerifier sets the signer and verifier. The newest certificate chain version
// in the store is used.
func setDefaultSignerVerifier(c *config.State, pubIA addr.IA) error {
	ctx, cancelF := context.WithTimeout(context.Background(), time.Second)
	defer cancelF()
	meta, err := trust.CreateSignMeta(ctx, pubIA, c.TrustDB)
	if err != nil {
		return err
	}
	signer, err := trust.NewBasicSigner(c.GetSigningKey(), meta)
	if err != nil {
		return err
	}
	c.SetSigner(signer)
	c.SetVerifier(c.Store.NewVerifier())
	return nil
}

// setMessenger sets the messenger and the internal messenger of the store in
// cfg.CS. This function may only be called once per config.
func setMessenger(cfg *config.Config, router snet.Router) error {
	topo := itopo.Get()
	topoAddress := topo.CS.GetById(cfg.General.ID)
	if topoAddress == nil {
		return common.NewBasicError("Unable to find topo address", nil)
	}
	nc := infraenv.NetworkConfig{
		IA:                    topo.ISD_AS,
		Public:                env.GetPublicSnetAddress(topo.ISD_AS, topoAddress),
		Bind:                  env.GetBindSnetAddress(topo.ISD_AS, topoAddress),
		SVC:                   addr.SvcCS,
		ReconnectToDispatcher: cfg.General.ReconnectToDispatcher,
		QUIC: infraenv.QUIC{
			Address:  cfg.QUIC.Address,
			CertFile: cfg.QUIC.CertFile,
			KeyFile:  cfg.QUIC.KeyFile,
		},
		SVCResolutionFraction: cfg.QUIC.ResolutionFraction,
		TrustStore:            state.Store,
		Router:                router,
		SVCRouter:             messenger.NewSVCRouter(itopo.Provider()),
	}
	var err error
	msgr, err = nc.Messenger()
	if err != nil {
		return common.NewBasicError("Unable to initialize SCION Messenger", err)
	}
	if cfg.DRKey.Enabled() {
		state.DRKeyStore.SetMessenger(msgr)
	}
	return nil
}

func addHandlers() error {
	topo := itopo.Get()
	msgr.AddHandler(infra.ChainRequest, state.Store.NewChainReqHandler(true))
	msgr.AddHandler(infra.TRCRequest, state.Store.NewTRCReqHandler(true))
	msgr.AddHandler(infra.Chain, state.Store.NewChainPushHandler())
	msgr.AddHandler(infra.TRC, state.Store.NewTRCPushHandler())
	signingTypes := []infra.MessageType{infra.ChainIssueRequest}
	if cfg.DRKey.Enabled() {
		signingTypes = []infra.MessageType{
			infra.DRKeyLvl1Request,
			infra.DRKeyLvl1Reply,
			infra.DRKeyLvl2Request,
		}
		msgr.AddHandler(infra.DRKeyLvl1Request, state.DRKeyStore.NewLvl1ReqHandler())
		msgr.AddHandler(infra.DRKeyLvl2Request, state.DRKeyStore.NewLvl2ReqHandler())
	}
	msgr.UpdateSigner(state.GetSigner(), signingTypes)
	msgr.UpdateVerifier(state.GetVerifier())
	// Only core CS handles certificate reissuance requests.
	if topo.Core {
		msgr.AddHandler(infra.ChainIssueRequest, &reiss.Handler{
			State: state,
			IA:    topo.ISD_AS,
		})
	}
	return nil
}
