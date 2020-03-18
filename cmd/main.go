package main

import (
	"github.com/xcoulon/kubectl-terminate/cmd/terminate"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE
)

func main() {
	terminate.InitAndExecute()
}
