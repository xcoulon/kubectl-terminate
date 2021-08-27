package main

import (
	"github.com/xcoulon/kubectl-terminate/cmd/terminate"
	_ "k8s.io/client-go/plugin/pkg/client/auth"     // see https://krew.sigs.k8s.io/docs/developer-guide/develop/best-practices/
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE
)

func main() {
	terminate.InitAndExecute()
}
