package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/akram/kgb/internal/scheme"
	"github.com/akram/kgb/pkg/audit"
	"github.com/akram/kgb/pkg/gateway"
	"github.com/akram/kgb/pkg/intent"
	"github.com/akram/kgb/pkg/policyengine"
	"github.com/akram/kgb/pkg/trustmode"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	var bind string
	var upstream string
	flag.StringVar(&bind, "bind", ":8080", "HTTP listen address.")
	flag.StringVar(&upstream, "upstream", "", "Optional upstream base URL for /proxy/* reverse proxy.")
	flag.Parse()

	cfg, err := ctrl.GetConfig()
	if err != nil {
		log.Fatalf("kubeconfig: %v", err)
	}
	cli, err := client.New(cfg, client.Options{Scheme: scheme.New()})
	if err != nil {
		log.Fatalf("client: %v", err)
	}

	srv := &gateway.Server{
		Client:       cli,
		Analyzer:     &intent.Chain{Steps: []intent.Analyzer{intent.RuleAnalyzer{}}},
		Engine:       &policyengine.Engine{Client: cli},
		Audit:        audit.NewLogger(),
		Approvals:    gateway.NewMemoryApprovals(),
		DefaultNS:    os.Getenv("KGB_DEFAULT_NAMESPACE"),
		TrustDefault: trustmode.FromEnv(),
	}
	if upstream != "" {
		pu, err := url.Parse(upstream)
		if err != nil {
			log.Fatalf("upstream: %v", err)
		}
		srv.UpstreamURL = pu
	}

	log.Printf("kgb-gateway listening on %s", bind)
	if err := http.ListenAndServe(bind, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
