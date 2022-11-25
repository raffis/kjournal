package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/server"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/util/feature"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"

	configv1alpha1 "github.com/raffis/kjournal/pkg/apis/config/v1alpha1"
	"github.com/raffis/kjournal/pkg/apiserver"
	"github.com/raffis/kjournal/pkg/storage"
)

var (
	schemeBuilder        runtime.SchemeBuilder
	storageProvider      map[schema.GroupResource]*storage.SingletonProvider
	groupVersions        map[schema.GroupVersion]bool
	orderedGroupVersions []schema.GroupVersion
)

var (
	provider storage.Provider
)

func init() {
	storageProvider = make(map[schema.GroupResource]*storage.SingletonProvider)
}

func withResourceAndHandler(obj resource.Object, sp apiserver.StorageProvider) {
	gvr := obj.GetGroupVersionResource()
	schemeBuilder.Register(resource.AddToScheme(obj))

	forGroupVersionResource(gvr, sp)
}

// forGroupVersionResource manually registers storage for a specific resource.
func forGroupVersionResource(
	gvr schema.GroupVersionResource, sp apiserver.StorageProvider) {
	// register the group version
	withGroupVersions(gvr.GroupVersion())

	if _, found := storageProvider[gvr.GroupResource()]; !found {
		storageProvider[gvr.GroupResource()] = &storage.SingletonProvider{Provider: sp}
	}

	// add the API with its storageProvider
	apiserver.APIs[gvr] = sp
}

func withGroupVersions(versions ...schema.GroupVersion) {
	if groupVersions == nil {
		groupVersions = map[schema.GroupVersion]bool{}
	}
	for _, gv := range versions {
		if _, found := groupVersions[gv]; found {
			continue
		}
		groupVersions[gv] = true
		orderedGroupVersions = append(orderedGroupVersions, gv)
	}
}

// ServerOptions contains state for master/api server
type ServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions

	StdOut io.Writer
	StdErr io.Writer
}

// NewServerOptions returns a new ServerOptions
func NewServerOptions(out, errOut io.Writer, versions ...schema.GroupVersion) *ServerOptions {

	sso := genericoptions.NewSecureServingOptions()

	// We are composing recommended options for an aggregated api-server,
	// whose client is typically a proxy multiplexing many operations ---
	// notably including long-running ones --- into one HTTP/2 connection
	// into this server.  So allow many concurrent operations.
	sso.HTTP2MaxStreamsPerConnection = 1000

	opts := &genericoptions.RecommendedOptions{
		SecureServing:              sso.WithLoopback(),
		Authentication:             genericoptions.NewDelegatingAuthenticationOptions(),
		Authorization:              genericoptions.NewDelegatingAuthorizationOptions(),
		Audit:                      genericoptions.NewAuditOptions(),
		Features:                   genericoptions.NewFeatureOptions(),
		CoreAPI:                    genericoptions.NewCoreAPIOptions(),
		FeatureGate:                feature.DefaultFeatureGate,
		ExtraAdmissionInitializers: func(c *server.RecommendedConfig) ([]admission.PluginInitializer, error) { return nil, nil },
		Admission:                  genericoptions.NewAdmissionOptions(),
		EgressSelector:             genericoptions.NewEgressSelectorOptions(),
		Traces:                     genericoptions.NewTracingOptions(),
	}

	o := &ServerOptions{
		RecommendedOptions: opts,
		StdOut:             out,
		StdErr:             errOut,
	}

	return o
}

// NewCommandStartServer provides a CLI handler for 'start master' command
// with a default ServerOptions.
func NewCommandStartServer(defaults *ServerOptions, stopCh <-chan struct{}) *cobra.Command {
	o := *defaults
	cmd := &cobra.Command{
		Short: "kjournal-apiserver",
		Long:  "Launches the kjournal kubernetes apiserver",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Validate(args); err != nil {
				return err
			}

			conf, err := initConfig()
			if err != nil {
				return err
			}

			pr, err := storage.NewProvider(conf)
			if err != nil {
				return err
			}

			provider = pr

			if err := o.RunServer(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	o.RecommendedOptions.AddFlags(flags)
	utilfeature.DefaultMutableFeatureGate.AddFlag(flags)

	return cmd
}

func initConfig() (conf configv1alpha1.APIServerConfig, err error) {
	b, err := ioutil.ReadFile("/config.yaml")
	if err != nil {
		return conf, fmt.Errorf("failed to read apiserver config: %w", err)
	}

	expand := os.ExpandEnv(string(b))

	scheme := runtime.NewScheme()
	_ = configv1alpha1.AddToScheme(scheme)
	codec := serializer.NewCodecFactory(scheme)
	decoder := codec.UniversalDeserializer()

	_, _, err = decoder.Decode([]byte(expand), nil, &conf)
	if err != nil {
		return conf, fmt.Errorf("failed to decode apiserver config: %w", err)
	}

	return conf, nil
}

// Validate validates ServerOptions
func (o ServerOptions) Validate(args []string) error {
	errors := []error{}
	errors = append(errors, o.RecommendedOptions.Validate()...)
	return utilerrors.NewAggregate(errors)
}

// Config returns config for the api server given ServerOptions
func (o *ServerOptions) Config() (*apiserver.Config, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)

	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig:   apiserver.ExtraConfig{},
	}
	return config, nil
}

// RunServer starts a new Server given ServerOptions
func (o ServerOptions) RunServer(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	server.GenericAPIServer.AddPostStartHookOrDie("start-server-informers", func(context genericapiserver.PostStartHookContext) error {
		if config.GenericConfig.SharedInformerFactory != nil {
			config.GenericConfig.SharedInformerFactory.Start(context.StopCh)
		}
		return nil
	})

	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}
