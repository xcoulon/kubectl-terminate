package main

import (
	"github.com/xcoulon/kubectl-terminate/cmd/terminate"
	_ "k8s.io/client-go/plugin/pkg/client/auth"     // see https://krew.sigs.k8s.io/docs/developer-guide/develop/best-practices/
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE
)

var (
	// BuildCommit lastest build commit (set by Makefile)
	BuildCommit = ""
	// BuildTag if the `BuildCommit` matches a tag
	BuildTag = ""
	// BuildTime set by build script (set by Makefile)
	BuildTime = ""
)

func main() {
	terminate.InitAndExecute()
}
