package orchestrator

type Orchestrator struct {
	RosImg    string
	RosHDD    string
	Iface     string
	IfaceCIDR string
	Cattle    string
	Reg       string
	Agent     string
}

func (o *Orchestrator) CreateAndBootstrap() error {
	return startAndRegisterVM(o.RosImg, o.RosHDD, o.Iface, o.IfaceCIDR, o.Cattle, o.Reg, o.Agent)
}

func (o *Orchestrator) DeleteVM() error {
	return nil
}
