package orchestrator

type Orchestrator struct {
	RosImg    string
	RosHDD    string
	Iface     string
	IfaceCIDR string
	ImageTag  string
	RegUrl    string
	ImageRepo string
	HostUuid  string
}

func (o *Orchestrator) CreateAndBootstrap() error {
	return startAndRegisterVM(o.RosImg, o.RosHDD, o.Iface, o.IfaceCIDR, o.ImageTag, o.ImageRepo, o.RegUrl, o.HostUuid)
}

func (o *Orchestrator) DeleteVM() error {
	return nil
}
