package executor

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/mesos/mesos-go/executor"
	"github.com/mesos/mesos-go/mesosproto"
	"github.com/rancher/rancher-mesos-executor/orchestrator"
)

type RancherExecutor struct {
	rosImg, iface, ifaceCIDR, rosHDD string
}

type TaskStruct struct {
	RegistrationUrl string `json: "registration_url"`
	ImageRepo       string `json: "image_repo"`
	ImageTag        string `json: "image_tag"`
	HostUuid        string `json: "host_uuid"`
	Hostname        string `json: "hostname"`
}

func NewRancherExecutor(imagePath, iface, ifaceCIDR, rosHDD string) *RancherExecutor {
	return &RancherExecutor{
		rosImg:    imagePath,
		iface:     iface,
		ifaceCIDR: ifaceCIDR,
		rosHDD:    rosHDD,
	}
}

func (e *RancherExecutor) Registered(executor.ExecutorDriver, *mesosproto.ExecutorInfo, *mesosproto.FrameworkInfo, *mesosproto.SlaveInfo) {
	log.Info("Registered Executor")
}

func (e *RancherExecutor) Reregistered(executor.ExecutorDriver, *mesosproto.SlaveInfo) {
	log.Info("Reregistered executor")
}

func (e *RancherExecutor) Disconnected(executor.ExecutorDriver) {
	log.Info("disconnected executor")
}

func (e *RancherExecutor) LaunchTask(d executor.ExecutorDriver, task *mesosproto.TaskInfo) {
	taskId := task.TaskId
	s := mesosproto.TaskState_TASK_RUNNING
	d.SendStatusUpdate(&mesosproto.TaskStatus{TaskId: taskId, State: &s})
	var taskStruct TaskStruct
	json.Unmarshal(task.Data, &taskStruct)
	orchestrator := &orchestrator.Orchestrator{
		RosImg:    e.rosImg,
		RosHDD:    e.rosHDD,
		Iface:     e.iface,
		IfaceCIDR: e.ifaceCIDR,
		ImageRepo: taskStruct.ImageRepo,
		RegUrl:    taskStruct.RegistrationUrl,
		ImageTag:  taskStruct.ImageTag,
		HostUuid:  taskStruct.HostUuid,
		Hostname:  taskStruct.Hostname,
	}
	err := orchestrator.CreateAndBootstrap()
	//TBD: Read message type and add supp for DELETE
	s = mesosproto.TaskState_TASK_RUNNING
	if err != nil {
		log.Error(err)
		s = mesosproto.TaskState_TASK_ERROR
	}
	d.SendStatusUpdate(&mesosproto.TaskStatus{TaskId: taskId, State: &s})
}

func (e *RancherExecutor) KillTask(executor.ExecutorDriver, *mesosproto.TaskID) {
	log.Info("killing task")
}

func (e *RancherExecutor) FrameworkMessage(_ executor.ExecutorDriver, message string) {
	log.WithFields(log.Fields{
		"msg": message,
	}).Info("Message received from framework")
}

func (e *RancherExecutor) Shutdown(executor.ExecutorDriver) {
	log.Info("Shutting down executor")
}

func (e *RancherExecutor) Error(executor.ExecutorDriver, string) {
	log.Info("Error while running executor")
}
